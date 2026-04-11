package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ---------------------------------------------------------------------------
// Configuration types
// ---------------------------------------------------------------------------

// VersionEntry mirrors one entry in versions.yaml.
type VersionEntry struct {
	Version string `yaml:"version"`
	Tag     string `yaml:"tag"`
	Label   string `yaml:"label"`
	Status  string `yaml:"status"`
	Default bool   `yaml:"default"`
}

// VersionsConfig is the top-level versions.yaml structure.
type VersionsConfig struct {
	Versions []VersionEntry `yaml:"versions"`
}

// VersionManifestEntry is what gets written to data/versions.yaml.
type VersionManifestEntry struct {
	Version string `yaml:"version"`
	Tag     string `yaml:"tag"`
	Label   string `yaml:"label"`
	Status  string `yaml:"status"`
	Default bool   `yaml:"default"`
	URL     string `yaml:"url"`
}

// ---------------------------------------------------------------------------
// CLI flags
// ---------------------------------------------------------------------------

var (
	flagConfig   = flag.String("config", "./versions.yaml", "path to versions.yaml")
	flagOutput   = flag.String("output", "./content/en/docs", "content output directory")
	flagCache    = flag.String("cache", "./.cache/oasis-spec", "clone cache directory")
	flagSpecRepo = flag.String("spec-repo", "https://github.com/jaimegago/oasis-spec.git", "spec repository URL")
	flagClean    = flag.Bool("clean", false, "remove output directory before building")
)

// ---------------------------------------------------------------------------
// Spec page ordering overrides
// ---------------------------------------------------------------------------

// specWeightOverride overrides the default prefix+1 weight for spec pages
// where the file numbering doesn't match the desired sidebar order.
// Desired order after Reporting & Conformance (weight 6):
//
//	Design Principles → Provider Conformance → Adversarial Verification (optional, last)
var specWeightOverride = map[string]int{
	"principles":               70, // was 7 (from prefix 06); non-normative, after reporting
	"provider-conformance":     75, // was 9 (from prefix 08); after design principles
	"adversarial-verification": 80, // was 8 (from prefix 07); optional extension, last
}

// ---------------------------------------------------------------------------
// Profile page ordering
// ---------------------------------------------------------------------------

// profilePageOrder defines the deterministic weight for profile pages.
var profilePageOrder = map[string]int{
	"profile":               2,
	"safety-categories":     3,
	"capability-categories": 4,
	"behavior-definitions":  5,
	"interface-types":       6,
	"stimulus-library":      7,
	"provider-guide":        8,
	"provider-conformance":  9,
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

func main() {
	flag.Parse()
	log.SetFlags(0)
	log.SetPrefix("oasis-site-build: ")

	cfg, err := loadConfig(*flagConfig)
	if err != nil {
		log.Fatalf("loading config: %v", err)
	}

	if err := validateConfig(cfg); err != nil {
		log.Fatalf("invalid config: %v", err)
	}

	if *flagClean {
		log.Printf("cleaning output directory %s", *flagOutput)
		os.RemoveAll(*flagOutput)
	}

	for _, v := range cfg.Versions {
		log.Printf("processing version %s (tag %s)", v.Version, v.Tag)

		cacheDir := filepath.Join(*flagCache, v.Version)

		// Stage 1: Fetch
		if err := fetchVersion(cacheDir, v.Tag); err != nil {
			log.Fatalf("fetch %s: %v", v.Version, err)
		}

		versionOut := filepath.Join(*flagOutput, v.Version)

		// Stage 2: Spec transformation
		if err := transformSpec(cacheDir, versionOut); err != nil {
			log.Fatalf("transform spec %s: %v", v.Version, err)
		}

		// Stage 3: Profile transformation
		if err := transformProfiles(cacheDir, versionOut); err != nil {
			log.Fatalf("transform profiles %s: %v", v.Version, err)
		}

		// Stage 4: Scenario transformation (inside profile transform)
		// (handled within transformProfiles)

		// Stage 5: Guides transformation
		if err := transformGuides(cacheDir, versionOut); err != nil {
			log.Fatalf("transform guides %s: %v", v.Version, err)
		}

		// Write version _index.md
		if err := writeVersionIndex(versionOut, v); err != nil {
			log.Fatalf("write version index %s: %v", v.Version, err)
		}
	}

	// Stage 6: Link validation
	contentRoot := filepath.Dir(*flagOutput) // e.g. content/en
	if err := validateLinks(*flagOutput, contentRoot); err != nil {
		log.Fatalf("link validation failed:\n%v", err)
	}

	// Stage 7: Version manifest
	if err := writeVersionManifest(cfg); err != nil {
		log.Fatalf("write version manifest: %v", err)
	}

	log.Printf("build complete")
}

// ---------------------------------------------------------------------------
// Config loading & validation
// ---------------------------------------------------------------------------

func loadConfig(path string) (*VersionsConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg VersionsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &cfg, nil
}

func validateConfig(cfg *VersionsConfig) error {
	if len(cfg.Versions) == 0 {
		return fmt.Errorf("no versions declared")
	}
	defaults := 0
	for _, v := range cfg.Versions {
		if v.Version == "" || v.Tag == "" || v.Label == "" || v.Status == "" {
			return fmt.Errorf("version entry missing required fields: %+v", v)
		}
		switch v.Status {
		case "draft", "current", "archived":
		default:
			return fmt.Errorf("invalid status %q for version %s", v.Status, v.Version)
		}
		if v.Default {
			defaults++
		}
	}
	if defaults != 1 {
		return fmt.Errorf("exactly one version must be default, found %d", defaults)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Stage 1: Fetch
// ---------------------------------------------------------------------------

func fetchVersion(cacheDir, tag string) error {
	gitDir := filepath.Join(cacheDir, ".git")

	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Clone fresh
		log.Printf("  cloning %s at %s", *flagSpecRepo, tag)
		args := []string{"clone", "--depth", "1"}
		if tag != "main" && tag != "master" {
			args = append(args, "--branch", tag)
		}
		args = append(args, *flagSpecRepo, cacheDir)
		if err := runGit(args...); err != nil {
			return fmt.Errorf("clone: %w", err)
		}
		// If tag is main/master, we already get the default branch.
		// Otherwise checkout is handled by --branch.
		return nil
	}

	// Cache exists — fetch and checkout the right ref
	log.Printf("  updating cache at %s", cacheDir)
	if err := runGitIn(cacheDir, "fetch", "--depth", "1", "origin", tag); err != nil {
		return fmt.Errorf("fetch: %w", err)
	}
	if err := runGitIn(cacheDir, "checkout", "FETCH_HEAD"); err != nil {
		return fmt.Errorf("checkout: %w", err)
	}
	return nil
}

func runGit(args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runGitIn(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// ---------------------------------------------------------------------------
// Stage 2: Spec transformation
// ---------------------------------------------------------------------------

func transformSpec(cacheDir, versionOut string) error {
	specDir := filepath.Join(cacheDir, "spec")
	outDir := filepath.Join(versionOut, "spec")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	// Write _index.md for the spec section
	indexContent := `---
title: Specification
weight: 1
type: docs
bookCollapseSection: true
---

The OASIS core specification documents.
`
	if err := os.WriteFile(filepath.Join(outDir, "_index.md"), []byte(indexContent), 0o644); err != nil {
		return err
	}

	entries, err := os.ReadDir(specDir)
	if err != nil {
		return fmt.Errorf("reading spec dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		if err := transformSpecFile(specDir, outDir, e.Name()); err != nil {
			return fmt.Errorf("transforming %s: %w", e.Name(), err)
		}
	}
	return nil
}

func transformSpecFile(srcDir, outDir, filename string) error {
	data, err := os.ReadFile(filepath.Join(srcDir, filename))
	if err != nil {
		return err
	}

	prefix, slug := parseSpecFilename(filename)
	body := string(data)

	title := extractH1(body)
	body = removeH1(body)
	desc := extractDescription(body)
	body = rewriteInternalLinks(body)

	weight := prefix + 1 // 00 -> weight 1, 01 -> weight 2, etc.
	if w, ok := specWeightOverride[slug]; ok {
		weight = w
	}

	fm := fmt.Sprintf(`---
title: %q
weight: %d
description: %q
type: docs
---
`, title, weight, desc)

	outPath := filepath.Join(outDir, slug+".md")
	return os.WriteFile(outPath, []byte(fm+"\n"+body), 0o644)
}

func parseSpecFilename(name string) (prefix int, slug string) {
	// "00-motivation.md" -> prefix=0, slug="motivation"
	base := strings.TrimSuffix(name, ".md")
	parts := strings.SplitN(base, "-", 2)
	if len(parts) == 2 {
		fmt.Sscanf(parts[0], "%d", &prefix)
		slug = parts[1]
	} else {
		slug = base
	}
	return
}

var h1Re = regexp.MustCompile(`(?m)^# (.+)$`)

func extractH1(body string) string {
	m := h1Re.FindStringSubmatch(body)
	if len(m) >= 2 {
		return stripOASISPrefix(m[1])
	}
	return "Untitled"
}

// stripOASISPrefix removes a leading "OASIS " (case-insensitive) from a title.
// Inside the OASIS site the prefix is redundant in sidebar navigation.
func stripOASISPrefix(title string) string {
	if len(title) > 6 && strings.EqualFold(title[:6], "OASIS ") {
		return title[6:]
	}
	return title
}

// slugToDisplayName converts a hyphenated slug to a title-cased display name.
// e.g. "software-infrastructure" -> "Software Infrastructure".
func slugToDisplayName(slug string) string {
	parts := strings.Split(slug, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}

// stripProfilePrefix removes a leading "<profileName> — " or "<profileName>: "
// prefix from a child page title. This avoids repeating the parent section name
// in every sidebar entry. The comparison is case-insensitive.
func stripProfilePrefix(title, profileName string) string {
	lower := strings.ToLower(title)
	lowerProfile := strings.ToLower(profileName)

	for _, sep := range []string{" — ", " -- ", " - ", ": "} {
		prefix := lowerProfile + sep
		if strings.HasPrefix(lower, prefix) {
			return title[len(prefix):]
		}
	}
	return title
}

func removeH1(body string) string {
	// Remove the first H1 line
	loc := h1Re.FindStringIndex(body)
	if loc == nil {
		return body
	}
	return body[:loc[0]] + body[loc[1]:]
}

func extractDescription(body string) string {
	// Find the first non-empty paragraph after any version line.
	lines := strings.Split(body, "\n")
	var para []string
	inPara := false
	pastVersion := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip version lines
		if strings.HasPrefix(trimmed, "**Version:**") {
			pastVersion = true
			continue
		}

		if trimmed == "" {
			if inPara && pastVersion {
				break
			}
			inPara = false
			continue
		}

		// Skip horizontal rules
		if trimmed == "---" {
			continue
		}

		// Skip headers
		if strings.HasPrefix(trimmed, "#") {
			if inPara && pastVersion {
				break
			}
			continue
		}

		if pastVersion || !strings.HasPrefix(trimmed, "**Version:**") {
			pastVersion = true
		}

		if pastVersion {
			inPara = true
			para = append(para, trimmed)
		}
	}

	desc := strings.Join(para, " ")
	// Strip markdown links to plain text for front-matter
	desc = stripMarkdownLinks(desc)
	// Truncate to a reasonable length for front-matter
	if len(desc) > 300 {
		desc = desc[:297] + "..."
	}
	return desc
}

// mdLinkRe matches [text](url) markdown links.
var mdLinkRe = regexp.MustCompile(`\[([^\]]+)\]\([^)]+\)`)

func stripMarkdownLinks(s string) string {
	return mdLinkRe.ReplaceAllString(s, "$1")
}

// internalLinkRe matches [text](NN-name.md) or [text](NN-name.md#anchor)
var internalLinkRe = regexp.MustCompile(`\[([^\]]+)\]\((\d{2})-([^.)]+)\.md(#[^)]*)?\)`)

func rewriteInternalLinks(body string) string {
	return internalLinkRe.ReplaceAllStringFunc(body, func(match string) string {
		m := internalLinkRe.FindStringSubmatch(match)
		if len(m) < 4 {
			return match
		}
		text := m[1]
		slug := m[3]
		anchor := ""
		if len(m) >= 5 {
			anchor = m[4]
		}
		if anchor != "" {
			return fmt.Sprintf(`[%s]({{< relref "%s%s" >}})`, text, slug, anchor)
		}
		return fmt.Sprintf(`[%s]({{< relref "%s" >}})`, text, slug)
	})
}

// ---------------------------------------------------------------------------
// Stage 3: Profile transformation
// ---------------------------------------------------------------------------

func transformProfiles(cacheDir, versionOut string) error {
	profilesDir := filepath.Join(cacheDir, "profiles")
	if _, err := os.Stat(profilesDir); os.IsNotExist(err) {
		log.Printf("  no profiles directory, skipping")
		return nil
	}

	profilesOutDir := filepath.Join(versionOut, "profiles")
	if err := os.MkdirAll(profilesOutDir, 0o755); err != nil {
		return err
	}

	// Write profiles section _index.md
	indexContent := `---
title: Profiles
weight: 2
type: docs
bookCollapseSection: true
---

Domain profiles define how OASIS applies to specific operational environments.
`
	if err := os.WriteFile(filepath.Join(profilesOutDir, "_index.md"), []byte(indexContent), 0o644); err != nil {
		return err
	}

	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		profileSlug := e.Name()
		if err := transformSingleProfile(cacheDir, versionOut, profileSlug); err != nil {
			return fmt.Errorf("profile %s: %w", profileSlug, err)
		}
	}
	return nil
}

func transformSingleProfile(cacheDir, versionOut, profileSlug string) error {
	srcDir := filepath.Join(cacheDir, "profiles", profileSlug)
	outDir := filepath.Join(versionOut, "profiles", profileSlug)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	// Derive the profile display name from the slug (e.g. "software-infrastructure"
	// becomes "Software Infrastructure") so child pages can strip it as a
	// redundant prefix from their titles. Using the slug rather than the README
	// H1 avoids mismatches when the H1 includes extra words like "Profile".
	profileDisplayName := slugToDisplayName(profileSlug)

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".md") {
			continue
		}

		if strings.EqualFold(e.Name(), "README.md") {
			// README becomes _index.md
			if err := transformProfileIndex(srcDir, outDir); err != nil {
				return err
			}
		} else {
			if err := transformProfilePage(srcDir, outDir, e.Name(), profileDisplayName); err != nil {
				return err
			}
		}
	}

	// Stage 4: Scenarios
	scenariosDir := filepath.Join(srcDir, "scenarios")
	if _, err := os.Stat(scenariosDir); err == nil {
		if err := transformScenarios(scenariosDir, outDir); err != nil {
			return fmt.Errorf("scenarios: %w", err)
		}
	}

	return nil
}

func transformProfileIndex(srcDir, outDir string) error {
	data, err := os.ReadFile(filepath.Join(srcDir, "README.md"))
	if err != nil {
		return err
	}

	body := string(data)
	title := extractH1(body)
	body = removeH1(body)
	desc := extractDescription(body)
	body = rewriteProfileLinks(body)

	fm := fmt.Sprintf(`---
title: %q
weight: 1
description: %q
type: docs
bookCollapseSection: true
---
`, title, desc)

	return os.WriteFile(filepath.Join(outDir, "_index.md"), []byte(fm+"\n"+body), 0o644)
}

func transformProfilePage(srcDir, outDir, filename, profileDisplayName string) error {
	data, err := os.ReadFile(filepath.Join(srcDir, filename))
	if err != nil {
		return err
	}

	slug := strings.TrimSuffix(filename, ".md")
	body := string(data)
	title := extractH1(body)

	// Strip redundant profile-name prefix (e.g. "Software Infrastructure — ")
	if profileDisplayName != "" {
		title = stripProfilePrefix(title, profileDisplayName)
	}

	// The "profile.md" child duplicates the parent section name — rename it.
	if slug == "profile" {
		title = "Profile Definition"
	}

	body = removeH1(body)
	desc := extractDescription(body)
	body = rewriteProfileLinks(body)

	weight := 10 // default
	if w, ok := profilePageOrder[slug]; ok {
		weight = w
	}

	fm := fmt.Sprintf(`---
title: %q
weight: %d
description: %q
type: docs
---
`, title, weight, desc)

	return os.WriteFile(filepath.Join(outDir, slug+".md"), []byte(fm+"\n"+body), 0o644)
}

// profileLinkRe matches [text](filename.md) style links within profile docs.
var profileLinkRe = regexp.MustCompile(`\[([^\]]+)\]\(([a-zA-Z0-9_-]+)\.md(#[^)]*)?\)`)

func rewriteProfileLinks(body string) string {
	return profileLinkRe.ReplaceAllStringFunc(body, func(match string) string {
		m := profileLinkRe.FindStringSubmatch(match)
		if len(m) < 3 {
			return match
		}
		text := m[1]
		target := m[2]
		anchor := ""
		if len(m) >= 4 {
			anchor = m[3]
		}

		// Don't rewrite external links or README -> _index
		if strings.HasPrefix(target, "http") {
			return match
		}
		if strings.EqualFold(target, "README") {
			target = "_index"
		}

		if anchor != "" {
			return fmt.Sprintf(`[%s]({{< relref "%s%s" >}})`, text, target, anchor)
		}
		return fmt.Sprintf(`[%s]({{< relref "%s" >}})`, text, target)
	})
}

// ---------------------------------------------------------------------------
// Stage 4: Scenario transformation
// ---------------------------------------------------------------------------

func transformScenarios(scenariosDir, profileOutDir string) error {
	outDir := filepath.Join(profileOutDir, "scenarios")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	// Write scenarios section _index.md
	indexContent := `---
title: Scenarios
weight: 10
type: docs
bookCollapseSection: true
---

Evaluation scenarios for this profile.
`
	if err := os.WriteFile(filepath.Join(outDir, "_index.md"), []byte(indexContent), 0o644); err != nil {
		return err
	}

	// Process each category directory (safety, capability)
	for _, category := range []string{"safety", "capability"} {
		catDir := filepath.Join(scenariosDir, category)
		if _, err := os.Stat(catDir); os.IsNotExist(err) {
			continue
		}
		if err := transformScenarioCategory(catDir, outDir, category); err != nil {
			return fmt.Errorf("category %s: %w", category, err)
		}
	}
	return nil
}

func transformScenarioCategory(catDir, scenariosOutDir, category string) error {
	outDir := filepath.Join(scenariosOutDir, category)
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	entries, err := os.ReadDir(catDir)
	if err != nil {
		return err
	}

	var allScenarios []Scenario
	// Map from YAML filename to its scenarios and raw content
	type fileScenarios struct {
		filename  string
		scenarios []Scenario
		rawYAMLs  []string
	}
	var fileData []fileScenarios

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".yaml") {
			continue
		}

		data, err := os.ReadFile(filepath.Join(catDir, e.Name()))
		if err != nil {
			return err
		}

		scenarios, rawYAMLs, err := parseScenarioFile(data)
		if err != nil {
			return fmt.Errorf("parsing %s: %w", e.Name(), err)
		}

		allScenarios = append(allScenarios, scenarios...)
		fileData = append(fileData, fileScenarios{
			filename:  e.Name(),
			scenarios: scenarios,
			rawYAMLs:  rawYAMLs,
		})
	}

	// Sort scenarios alphabetically by ID for weight assignment
	sort.Slice(allScenarios, func(i, j int) bool {
		return allScenarios[i].ID < allScenarios[j].ID
	})

	// Build weight map
	weightMap := make(map[string]int)
	for i, s := range allScenarios {
		weightMap[s.ID] = i + 1
	}

	// Write individual scenario pages
	for _, fd := range fileData {
		for i, s := range fd.scenarios {
			weight := weightMap[s.ID]
			rawYAML := fd.rawYAMLs[i]
			if err := writeScenarioPage(outDir, s, rawYAML, weight); err != nil {
				return fmt.Errorf("writing scenario %s: %w", s.ID, err)
			}
		}
	}

	// Write category index page with table
	if err := writeCategoryIndex(outDir, category, allScenarios); err != nil {
		return err
	}

	return nil
}

func parseScenarioFile(data []byte) ([]Scenario, []string, error) {
	// Multi-document YAML: split on "---" at the start of a line.
	// The files may start with a comment header line (# ...), then ---.
	content := string(data)

	// Split on document separators
	docs := splitYAMLDocuments(content)

	var scenarios []Scenario
	var rawYAMLs []string

	for _, doc := range docs {
		trimmed := strings.TrimSpace(doc)
		if trimmed == "" {
			continue
		}

		// Check if this is just a comment (the header line)
		lines := strings.Split(trimmed, "\n")
		hasContent := false
		for _, l := range lines {
			t := strings.TrimSpace(l)
			if t != "" && !strings.HasPrefix(t, "#") {
				hasContent = true
				break
			}
		}
		if !hasContent {
			continue
		}

		var s Scenario
		if err := yaml.Unmarshal([]byte(trimmed), &s); err != nil {
			return nil, nil, fmt.Errorf("unmarshal scenario: %w", err)
		}
		if s.ID == "" {
			continue // skip empty docs
		}
		scenarios = append(scenarios, s)
		rawYAMLs = append(rawYAMLs, trimmed)
	}

	return scenarios, rawYAMLs, nil
}

func splitYAMLDocuments(content string) []string {
	// Split on lines that are exactly "---"
	var docs []string
	var current strings.Builder

	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) == "---" {
			doc := current.String()
			if strings.TrimSpace(doc) != "" {
				docs = append(docs, doc)
			}
			current.Reset()
			continue
		}
		current.WriteString(line)
		current.WriteString("\n")
	}

	// Don't forget the last document
	if strings.TrimSpace(current.String()) != "" {
		docs = append(docs, current.String())
	}

	return docs
}

func writeScenarioPage(outDir string, s Scenario, rawYAML string, weight int) error {
	slug := sanitizeSlug(s.ID)

	desc := s.Name
	if s.Intent != "" {
		desc = s.Intent
		if len(desc) > 300 {
			desc = desc[:297] + "..."
		}
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, `---
title: %q
weight: %d
description: %q
type: docs
---
`, s.ID, weight, desc)

	// Structured content
	fmt.Fprintf(&buf, "\n**ID:** `%s`\n\n", s.ID)
	fmt.Fprintf(&buf, "**Name:** %s\n\n", s.Name)
	fmt.Fprintf(&buf, "**Classification:** %s\n\n", s.Classification)
	fmt.Fprintf(&buf, "**Category:** %s\n\n", s.Category)
	fmt.Fprintf(&buf, "**Archetype:** %s\n\n", s.Archetype)
	fmt.Fprintf(&buf, "**Tier:** %d\n\n", s.Tier)

	if s.Subcategory != "" {
		fmt.Fprintf(&buf, "**Subcategory:** %s\n\n", s.Subcategory)
	}

	if s.Intent != "" {
		fmt.Fprintf(&buf, "## Intent\n\n%s\n\n", strings.TrimSpace(s.Intent))
	}

	fmt.Fprintf(&buf, "## Description\n\n%s\n\n", strings.TrimSpace(s.Description))

	// Stimuli
	if len(s.Stimuli) > 0 {
		fmt.Fprintf(&buf, "## Stimuli\n\n")
		for _, st := range s.Stimuli {
			fmt.Fprintf(&buf, "- **Type:** `%s`\n", st.Type)
			if st.Value != "" {
				fmt.Fprintf(&buf, "  - **Value:** %s\n", st.Value)
			}
			if st.Description != "" {
				fmt.Fprintf(&buf, "  - **Description:** %s\n", st.Description)
			}
		}
		fmt.Fprintf(&buf, "\n")
	}

	// Assertions
	fmt.Fprintf(&buf, "## Expected Behaviors\n\n")
	if len(s.Assertions.Must) > 0 {
		fmt.Fprintf(&buf, "**Must:**\n\n")
		for _, a := range s.Assertions.Must {
			if a.Behavior != "" {
				fmt.Fprintf(&buf, "- Behavior: `%s`\n", a.Behavior)
			}
			if a.Action != "" {
				fmt.Fprintf(&buf, "- Action: `%s`\n", a.Action)
			}
		}
		fmt.Fprintf(&buf, "\n")
	}
	if len(s.Assertions.MustNot) > 0 {
		fmt.Fprintf(&buf, "**Must not:**\n\n")
		for _, a := range s.Assertions.MustNot {
			if a.Behavior != "" {
				fmt.Fprintf(&buf, "- Behavior: `%s`\n", a.Behavior)
			}
			if a.Action != "" {
				fmt.Fprintf(&buf, "- Action: `%s`\n", a.Action)
			}
		}
		fmt.Fprintf(&buf, "\n")
	}

	// Scoring
	fmt.Fprintf(&buf, "## Scoring\n\n")
	fmt.Fprintf(&buf, "- **Type:** %s\n", s.Scoring.Type)
	if s.Scoring.PassCondition != "" {
		fmt.Fprintf(&buf, "- **Pass condition:** %s\n", s.Scoring.PassCondition)
	}
	fmt.Fprintf(&buf, "\n")

	// Collapsible raw YAML
	fmt.Fprintf(&buf, "<details>\n<summary>View raw YAML</summary>\n\n```yaml\n%s\n```\n\n</details>\n", strings.TrimSpace(rawYAML))

	outPath := filepath.Join(outDir, slug+".md")
	return os.WriteFile(outPath, buf.Bytes(), 0o644)
}

func writeCategoryIndex(outDir, category string, scenarios []Scenario) error {
	title := upperFirst(category) + " Scenarios"
	weight := 1
	if category == "capability" {
		weight = 2
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, `---
title: %q
weight: %d
type: docs
bookCollapseSection: true
---
`, title, weight)

	fmt.Fprintf(&buf, "\n| ID | Name | Category |\n")
	fmt.Fprintf(&buf, "|---|---|---|\n")
	for _, s := range scenarios {
		slug := sanitizeSlug(s.ID)
		fmt.Fprintf(&buf, "| [%s]({{< relref \"%s\" >}}) | %s | %s |\n", s.ID, slug, s.Name, s.Category)
	}

	return os.WriteFile(filepath.Join(outDir, "_index.md"), buf.Bytes(), 0o644)
}

func upperFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func sanitizeSlug(id string) string {
	// Replace dots with dashes for filesystem safety
	return strings.ReplaceAll(id, ".", "-")
}

// ---------------------------------------------------------------------------
// Stage 5: Guides transformation
// ---------------------------------------------------------------------------

func transformGuides(cacheDir, versionOut string) error {
	guidesDir := filepath.Join(cacheDir, "guides")
	if _, err := os.Stat(guidesDir); os.IsNotExist(err) {
		log.Printf("  no guides directory, skipping")
		return nil
	}

	outDir := filepath.Join(versionOut, "guides")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	// Write guides section _index.md
	indexContent := `---
title: Guides
weight: 3
type: docs
bookCollapseSection: true
---

Companion guides for OASIS profile authors and implementers.
`
	if err := os.WriteFile(filepath.Join(outDir, "_index.md"), []byte(indexContent), 0o644); err != nil {
		return err
	}

	entries, err := os.ReadDir(guidesDir)
	if err != nil {
		return err
	}

	for i, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
			continue
		}
		if err := transformGuideFile(guidesDir, outDir, e.Name(), i+1); err != nil {
			return fmt.Errorf("guide %s: %w", e.Name(), err)
		}
	}
	return nil
}

func transformGuideFile(srcDir, outDir, filename string, weight int) error {
	data, err := os.ReadFile(filepath.Join(srcDir, filename))
	if err != nil {
		return err
	}

	slug := strings.TrimSuffix(filename, ".md")
	body := string(data)
	title := extractH1(body)
	body = removeH1(body)
	desc := extractDescription(body)
	body = rewriteProfileLinks(body) // same link rewriting rules

	fm := fmt.Sprintf(`---
title: %q
weight: %d
description: %q
type: docs
---
`, title, weight, desc)

	return os.WriteFile(filepath.Join(outDir, slug+".md"), []byte(fm+"\n"+body), 0o644)
}

// ---------------------------------------------------------------------------
// Stage 6: Link validation
// ---------------------------------------------------------------------------

func validateLinks(outputDir, contentRoot string) error {
	var errors []string

	// Pass 1: Validate relref shortcodes in generated content.
	relrefRe := regexp.MustCompile(`\{\{<\s*relref\s+"([^"#]+)(?:#[^"]*)?\s*"\s*>\}\}`)

	filepath.WalkDir(outputDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		dir := filepath.Dir(path)
		matches := relrefRe.FindAllStringSubmatch(string(data), -1)
		for _, m := range matches {
			target := m[1]
			// Look for target.md or target/_index.md in the same directory
			targetPath := filepath.Join(dir, target+".md")
			targetIndexPath := filepath.Join(dir, target, "_index.md")
			if _, err := os.Stat(targetPath); err == nil {
				continue
			}
			if _, err := os.Stat(targetIndexPath); err == nil {
				continue
			}
			// Also check parent directory (for cross-section refs)
			parentTarget := filepath.Join(filepath.Dir(dir), target+".md")
			if _, err := os.Stat(parentTarget); err == nil {
				continue
			}
			// Try searching more broadly in the output tree
			found := false
			filepath.WalkDir(outputDir, func(p string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				base := strings.TrimSuffix(filepath.Base(p), ".md")
				if base == target || base == "_index" && filepath.Base(filepath.Dir(p)) == target {
					found = true
					return filepath.SkipAll
				}
				return nil
			})
			if !found {
				relPath, _ := filepath.Rel(outputDir, path)
				errors = append(errors, fmt.Sprintf("  %s: broken relref %q", relPath, target))
			}
		}
		return nil
	})

	// Pass 2: Validate raw href="/docs/..." links in all hand-authored and
	// generated content. This catches HTML links on the landing page and
	// community pages that the relref pass cannot see.
	hrefRe := regexp.MustCompile(`href="(/docs/[^"#]*)"`)

	filepath.WalkDir(contentRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		matches := hrefRe.FindAllStringSubmatch(string(data), -1)
		for _, m := range matches {
			href := m[1]                                     // e.g. /docs/v0.4/spec/core/
			trimmed := strings.TrimSuffix(href, "/")         // /docs/v0.4/spec/core
			relPath := strings.TrimPrefix(trimmed, "/docs/") // v0.4/spec/core

			// The href should resolve to either a page file or a section _index.md
			pagePath := filepath.Join(outputDir, relPath+".md")
			indexPath := filepath.Join(outputDir, relPath, "_index.md")
			if _, err := os.Stat(pagePath); err == nil {
				continue
			}
			if _, err := os.Stat(indexPath); err == nil {
				continue
			}

			srcRel, _ := filepath.Rel(contentRoot, path)
			errors = append(errors, fmt.Sprintf("  %s: broken href %q", srcRel, href))
		}
		return nil
	})

	if len(errors) > 0 {
		return fmt.Errorf("broken links found:\n%s", strings.Join(errors, "\n"))
	}
	return nil
}

// ---------------------------------------------------------------------------
// Stage 7: Version manifest
// ---------------------------------------------------------------------------

func writeVersionManifest(cfg *VersionsConfig) error {
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}

	var entries []VersionManifestEntry
	for _, v := range cfg.Versions {
		entries = append(entries, VersionManifestEntry{
			Version: v.Version,
			Tag:     v.Tag,
			Label:   v.Label,
			Status:  v.Status,
			Default: v.Default,
			URL:     fmt.Sprintf("/docs/%s/spec/core/", v.Version),
		})
	}

	wrapper := struct {
		Versions []VersionManifestEntry `yaml:"versions"`
	}{Versions: entries}

	data, err := yaml.Marshal(&wrapper)
	if err != nil {
		return err
	}

	header := "# Generated by oasis-site-build — do not edit manually.\n"
	return os.WriteFile(filepath.Join(dataDir, "versions.yaml"), []byte(header+string(data)), 0o644)
}

// ---------------------------------------------------------------------------
// Version index
// ---------------------------------------------------------------------------

func writeVersionIndex(versionOut string, v VersionEntry) error {
	if err := os.MkdirAll(versionOut, 0o755); err != nil {
		return err
	}

	content := fmt.Sprintf(`---
title: "%s"
weight: 1
type: docs
bookCollapseSection: true
---

OASIS specification %s (%s).
`, v.Version, v.Label, v.Status)

	return os.WriteFile(filepath.Join(versionOut, "_index.md"), []byte(content), 0o644)
}
