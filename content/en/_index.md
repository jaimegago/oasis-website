---
title: "OASIS — Open Assessment Standard for Intelligent Systems"
type: docs
---

<!-- Section 1: Hero -->
<div class="oasis-hero">
  <h1>OASIS</h1>
  <p class="oasis-hero-tagline">Open Assessment Standard for Intelligent Systems</p>
  <p class="oasis-hero-version"><a href="https://github.com/jaimegago/oasis-spec/blob/main/CHANGELOG.md" class="version-badge">v1.0.0-rc1</a></p>
  <div class="oasis-rc-banner">
    <p><strong>Release candidate:</strong> OASIS v1.0.0-rc1 is feature-complete and validated end-to-end against a real AI infrastructure agent. Seeking external feedback before v1.0.0 stability guarantees.</p>
    <p class="oasis-rc-banner-links"><a href="https://github.com/jaimegago/oasis-spec/blob/main/CHANGELOG.md">Changelog</a> · <a href="https://github.com/jaimegago/oasis-spec/issues">Open an issue</a></p>
  </div>
  <p class="oasis-hero-thesis">OASIS is an open standard for evaluating AI agents that operate in real-world systems. It defines how to test whether an agent is safe to deploy and how capable it is — in that order. Safety is a gate, not a score: an agent that fails any safety scenario receives no capability score at all.</p>
  <div class="oasis-hero-ctas">
    <a href="/docs/v1.0/spec/core/" class="book-btn">Read the spec</a>
    <a href="/docs/v1.0/profiles/software-infrastructure/" class="book-btn book-btn-secondary">Browse the Software Infrastructure profile</a>
  </div>
</div>

<!-- Section 2: How it works -->
<div class="oasis-section">
  <h2>How an OASIS evaluation runs</h2>
  <p>Every OASIS evaluation proceeds through three phases. The order is normative.</p>

  <div class="oasis-phases-diagram">
    <svg viewBox="0 0 780 140" xmlns="http://www.w3.org/2000/svg" role="img" aria-labelledby="oasis-phases-title oasis-phases-desc">
      <title id="oasis-phases-title">Three-phase OASIS evaluation flow: Provider Conformance Preflight, then Safety Gate, then Capability Scoring</title>
      <desc id="oasis-phases-desc">A horizontal flow diagram showing the three sequential phases of an OASIS evaluation. Phase 0, Provider Conformance Preflight, verifies the provider meets profile requirements. An arrow leads to Phase 1, Safety Gate, where any failure halts the evaluation. A second arrow leads to Phase 2, Capability Scoring, which only runs if the safety gate passes.</desc>
      <!-- Phase 0 -->
      <rect x="0" y="0" width="220" height="80" rx="6" ry="6" fill="var(--oasis-bg-elevated)" stroke="var(--oasis-border)" stroke-width="1"/>
      <rect x="0" y="0" width="220" height="4" rx="6" ry="6" fill="var(--oasis-accent)"/>
      <text x="110" y="45" text-anchor="middle" fill="var(--oasis-text)" font-size="14" font-weight="600" font-family="IBM Plex Sans, sans-serif">Provider Conformance</text>
      <text x="110" y="65" text-anchor="middle" fill="var(--oasis-text)" font-size="14" font-weight="600" font-family="IBM Plex Sans, sans-serif">Preflight</text>
      <text x="110" y="100" text-anchor="middle" fill="var(--oasis-text-secondary)" font-size="12" font-family="IBM Plex Sans, sans-serif"><tspan x="110" dy="0">Verify provider meets</tspan><tspan x="110" dy="1.1em">profile requirements</tspan></text>
      <!-- Arrow 1 -->
      <line x1="230" y1="40" x2="268" y2="40" stroke="var(--oasis-text-tertiary)" stroke-width="2"/>
      <polygon points="268,35 278,40 268,45" fill="var(--oasis-text-tertiary)"/>
      <!-- Phase 1 -->
      <rect x="280" y="0" width="220" height="80" rx="6" ry="6" fill="var(--oasis-bg-elevated)" stroke="var(--oasis-border)" stroke-width="1"/>
      <rect x="280" y="0" width="220" height="4" rx="6" ry="6" fill="var(--oasis-accent)"/>
      <text x="390" y="50" text-anchor="middle" fill="var(--oasis-text)" font-size="14" font-weight="600" font-family="IBM Plex Sans, sans-serif">Safety Gate</text>
      <text x="390" y="100" text-anchor="middle" fill="var(--oasis-text-secondary)" font-size="12" font-family="IBM Plex Sans, sans-serif"><tspan x="390" dy="0">Any failure halts</tspan><tspan x="390" dy="1.1em">the evaluation</tspan></text>
      <!-- Arrow 2 -->
      <line x1="510" y1="40" x2="548" y2="40" stroke="var(--oasis-text-tertiary)" stroke-width="2"/>
      <polygon points="548,35 558,40 548,45" fill="var(--oasis-text-tertiary)"/>
      <!-- Phase 2 -->
      <rect x="560" y="0" width="220" height="80" rx="6" ry="6" fill="var(--oasis-bg-elevated)" stroke="var(--oasis-border)" stroke-width="1"/>
      <rect x="560" y="0" width="220" height="4" rx="6" ry="6" fill="var(--oasis-accent)"/>
      <text x="670" y="50" text-anchor="middle" fill="var(--oasis-text)" font-size="14" font-weight="600" font-family="IBM Plex Sans, sans-serif">Capability Scoring</text>
      <text x="670" y="100" text-anchor="middle" fill="var(--oasis-text-secondary)" font-size="12" font-family="IBM Plex Sans, sans-serif"><tspan x="670" dy="0">Only runs if the</tspan><tspan x="670" dy="1.1em">safety gate passes</tspan></text>
    </svg>
  </div>

  <h3>Phase 0 — Provider Conformance Preflight</h3>
  <p>Before any scenario runs, the runner queries the provider's conformance endpoint and verifies that every requirement declared by the active profile is satisfied. A mismatch aborts the run before a single scenario executes. This guarantees that every reported result was produced under conditions the profile considers valid.</p>

  <h3>Phase 1 — Safety Gate</h3>
  <p>The runner executes every safety scenario in the profile and aggregates the verdicts. The gate is binary: any safety failure, above a profile-defined tolerance, halts the evaluation. No capability scenarios run. No capability score is produced. The report records the failure and stops.</p>

  <h3>Phase 2 — Capability Scoring</h3>
  <p>Only agents that clear the safety gate are scored on capability. The capability section of the report is meaningful precisely because it cannot exist without a passing safety gate above it.</p>
</div>

<!-- Section 3: Differentiators -->
<div class="oasis-section">
  <h2>What makes OASIS different</h2>
  <div class="oasis-differentiators">
    <div class="oasis-differentiator">
      <h3>Safety is a gate, not a score</h3>
      <p>Existing benchmarks average safety into capability, hiding catastrophic failures behind good aggregate numbers. OASIS refuses to score capability until safety passes completely. There is no weighted blend, no partial credit, no "mostly safe."</p>
    </div>
    <div class="oasis-differentiator">
      <h3>Domain profiles, not generic benchmarks</h3>
      <p>A spec defines structure; a profile defines what safe and capable mean for a specific domain. Software Infrastructure is the first profile. Generic benchmarks cannot capture domain-specific risk, and OASIS does not pretend they can.</p>
    </div>
    <div class="oasis-differentiator">
      <h3>Independent verification by design</h3>
      <p>Provider conformance is checked at runtime against a published contract. Adversarial verification is a first-class extension, not an afterthought. The standard is built to be audited by parties other than the agent's vendor.</p>
    </div>
  </div>
</div>

<!-- Section 4: Specification map -->
<div class="oasis-section">
  <h2>The specification</h2>
  <p>OASIS is nine documents. Seven are normative; two provide context.</p>
  <div class="oasis-spec-map">
    <div class="oasis-spec-map-column">
      <h3>Normative</h3>
      <ul>
        <li><a href="/docs/v1.0/spec/core/">Core</a> — Foundational concepts, evaluation model, and architecture.</li>
        <li><a href="/docs/v1.0/spec/scenarios/">Scenarios</a> — Schema for scenarios and scenario suites.</li>
        <li><a href="/docs/v1.0/spec/profiles/">Profiles</a> — Structure and quality criteria for domain profiles.</li>
        <li><a href="/docs/v1.0/spec/execution/">Execution</a> — Agent interface contract and environment model.</li>
        <li><a href="/docs/v1.0/spec/reporting/">Reporting &amp; Conformance</a> — Verdict format and report structure.</li>
        <li><a href="/docs/v1.0/spec/provider-conformance/">Provider Conformance</a> — Requirements for OASIS-conformant evaluation providers.</li>
        <li><a href="/docs/v1.0/spec/adversarial-verification/">Adversarial Verification</a> — Optional extension for non-deterministic adversarial testing.</li>
      </ul>
    </div>
    <div class="oasis-spec-map-column">
      <h3>Non-normative</h3>
      <ul>
        <li><a href="/docs/v1.0/spec/motivation/">Motivation</a> — Why OASIS exists and what gap it fills.</li>
        <li><a href="/docs/v1.0/spec/principles/">Design Principles</a> — The principles the standard is built on.</li>
      </ul>
    </div>
  </div>
</div>

<!-- Section 5: Validation -->
<div class="oasis-section">
  <h2>Validated end-to-end</h2>
  <p>The release candidate was validated by executing the full SI v0.2 profile against a real AI agent (Joe) operating on a live Kubernetes cluster — not a simulation or mock environment.</p>
  <div class="oasis-status-grid">
    <div class="oasis-status-item">
      <span class="oasis-status-label">Safety scenarios executed</span>
      <span class="oasis-status-value">21 — all with deterministic PASS/FAIL verdicts</span>
    </div>
    <div class="oasis-status-item">
      <span class="oasis-status-label">Provision failures</span>
      <span class="oasis-status-value">0 — every scenario provisioned and executed cleanly</span>
    </div>
    <div class="oasis-status-item">
      <span class="oasis-status-label">Missing-heuristic errors</span>
      <span class="oasis-status-value">0 — every assertion resolved to a definitive verdict</span>
    </div>
  </div>
  <p>This confirms the spec is implementable, not theoretical. The evaluation pipeline, scenario schema, and verdict model all function as specified under real-world conditions.</p>
</div>

<!-- Section 6: Status -->
<div class="oasis-section">
  <h2>Status</h2>
  <div class="oasis-status-grid">
    <div class="oasis-status-item">
      <span class="oasis-status-label">Current version</span>
      <span class="oasis-status-value">v1.0.0-rc1 — release candidate, under review</span>
    </div>
    <div class="oasis-status-item">
      <span class="oasis-status-label">Current profile</span>
      <span class="oasis-status-value">Software Infrastructure — 7 safety categories (21 scenario archetypes), 7 capability categories (30 scenario archetypes), every archetype mapped to a real infrastructure failure mode</span>
    </div>
    <div class="oasis-status-item">
      <span class="oasis-status-label">Reference tooling</span>
      <span class="oasis-status-value">oasisctl (runner), Petri (ephemeral Kubernetes test environments), and a reference adapter — all working end-to-end against the SI profile</span>
    </div>
    <div class="oasis-status-item">
      <span class="oasis-status-label">Next</span>
      <span class="oasis-status-value">v1.0.0 final after the RC feedback period · second domain profile · adversarial verification reference implementation</span>
    </div>
  </div>
  <p>OASIS is developed in the open. The specification, profiles, and tooling live at <a href="https://github.com/jaimegago/oasis-spec">github.com/jaimegago/oasis-spec</a>.</p>
</div>
