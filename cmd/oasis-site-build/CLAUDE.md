# oasis-site-build

Go command-line tool that reads the OASIS specification from tagged versions of the oasis-spec repository and generates Hugo-ready markdown content.

## Pipeline stages

1. **Fetch** — For each version in `versions.yaml`, clone or update a cached copy of the oasis-spec repository at the declared git tag. Cache lives in `.cache/oasis-spec/`.

2. **Spec transformation** — Walk `spec/` in the cached clone. For each numbered markdown file (e.g. `00-motivation.md`), produce a Hugo page under `content/en/docs/vX.Y/spec/<slug>.md`. The slug drops the numeric prefix. Front-matter is computed from the file content (title from H1, weight from prefix, description from first paragraph). The original H1 is removed. Internal links like `[text](NN-name.md)` are rewritten to `{{< relref "name" >}}`.

3. **Profile transformation** — Walk `profiles/` in the cached clone. For each profile directory, `README.md` becomes `_index.md` and each other markdown file becomes its own page with deterministic weight ordering (profile, safety-categories, capability-categories, behavior-definitions, interface-types, stimulus-library, provider-guide, provider-conformance).

4. **Scenario transformation** — For each profile, walk `scenarios/safety/` and `scenarios/capability/`. Parse multi-document YAML files into Go structs. Each scenario becomes a page with structured content (id, category, stimuli, assertions, scoring) followed by a collapsible raw YAML section. Category index pages render a table of all scenarios.

5. **Guides transformation** — Walk `guides/` in the cached clone. Each markdown file becomes a page under `content/en/docs/vX.Y/guides/`.

6. **Link validation** — Walk the output tree and verify every `relref` link resolves to an existing file. Hard failure on broken links.

7. **Version manifest** — Write `data/versions.yaml` for Hugo to render the version dropdown.

## Go files

| File | Purpose |
|------|---------|
| `main.go` | CLI entry point, all pipeline stages, content transformation logic |
| `scenario_schema.go` | Go struct definitions matching the OASIS scenario YAML schema |

## Critical invariant

**When the scenario schema in the spec repository changes, `scenario_schema.go` in this script MUST be updated to match in the same change.** This invariant cannot be enforced automatically.

The authoritative source for the scenario schema is `spec/02-scenarios.md` in the oasis-spec repository (github.com/jaimegago/oasis-spec).

## CLI flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `./versions.yaml` | Path to versions.yaml |
| `--output` | `./content/en/docs` | Content output directory |
| `--cache` | `./.cache/oasis-spec` | Clone cache directory |
| `--spec-repo` | `https://github.com/jaimegago/oasis-spec.git` | Spec repository URL |
| `--clean` | `false` | Remove output directory before building |
