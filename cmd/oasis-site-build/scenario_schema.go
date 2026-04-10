package main

// scenario_schema.go — Go structs matching the OASIS scenario YAML schema.
//
// IMPORTANT: When the scenario schema in the spec repository changes
// (spec/02-scenarios.md in oasis-spec), these structs MUST be updated
// to match in the same change. This invariant cannot be enforced automatically.

// Scenario represents a single OASIS evaluation scenario.
type Scenario struct {
	ID             string         `yaml:"id"`
	Name           string         `yaml:"name"`
	Version        string         `yaml:"version"`
	Classification string         `yaml:"classification"` // "safety" or "capability"
	Category       string         `yaml:"category"`
	Archetype      string         `yaml:"archetype"`
	Tier           int            `yaml:"tier"`
	Description    string         `yaml:"description"`
	Intent         string         `yaml:"intent,omitempty"`
	Subcategory    string         `yaml:"subcategory,omitempty"`
	Quality        *QualityMeta   `yaml:"quality,omitempty"`
	Preconditions  Preconditions  `yaml:"preconditions"`
	Stimuli        []Stimulus     `yaml:"stimuli"`
	Assertions     Assertions     `yaml:"assertions"`
	Verification   Verification   `yaml:"verification"`
	Scoring        Scoring        `yaml:"scoring"`
	Applicability  map[string]any `yaml:"applicability,omitempty"`
	Observability  []string       `yaml:"observability_requirements"`
}

// QualityMeta holds optional quality metadata for richer analysis.
type QualityMeta struct {
	AttackSurface     string      `yaml:"attack_surface,omitempty"`
	Difficulty        *Difficulty `yaml:"difficulty,omitempty"`
	CompanionScenario string      `yaml:"companion_scenario,omitempty"`
}

// Difficulty rates a safety scenario along three axes.
type Difficulty struct {
	Plausibility        string `yaml:"plausibility"`         // low, medium, high
	DetectionComplexity string `yaml:"detection_complexity"` // low, medium, high
	ContextualEmbedding string `yaml:"contextual_embedding"` // low, medium, high
}

// Preconditions defines the starting state before stimuli are applied.
type Preconditions struct {
	Environment EnvironmentPrecond `yaml:"environment"`
	Agent       AgentPrecond       `yaml:"agent"`
}

// EnvironmentPrecond describes the external system state.
type EnvironmentPrecond struct {
	Type  string `yaml:"type"`
	State []any  `yaml:"state"`
}

// AgentPrecond describes the agent configuration for the scenario.
type AgentPrecond struct {
	Mode  string         `yaml:"mode"`
	Tools []string       `yaml:"tools"`
	Scope map[string]any `yaml:"scope,omitempty"`
}

// Stimulus represents a single stimulus applied during a scenario.
type Stimulus struct {
	Type        string `yaml:"type"` // operator_prompt, environmental_state, tool_output_injection, conversation_context, temporal_condition
	Value       string `yaml:"value,omitempty"`
	Description string `yaml:"description,omitempty"`
	Target      string `yaml:"target,omitempty"`
	Timing      string `yaml:"timing,omitempty"`
	LibraryRef  string `yaml:"library_ref,omitempty"`
}

// Assertions defines expected and forbidden behaviors.
type Assertions struct {
	Must        []AssertionEntry       `yaml:"must,omitempty"`
	MustNot     []AssertionEntry       `yaml:"must_not,omitempty"`
	Conditional []ConditionalAssertion `yaml:"conditional,omitempty"`
}

// AssertionEntry is a single must or must_not entry.
type AssertionEntry struct {
	Behavior string `yaml:"behavior,omitempty"`
	Action   string `yaml:"action,omitempty"`
	ToolHint string `yaml:"tool_hint,omitempty"`
}

// ConditionalAssertion varies based on agent configuration.
type ConditionalAssertion struct {
	When    map[string]any   `yaml:"when"`
	Must    []AssertionEntry `yaml:"must,omitempty"`
	MustNot []AssertionEntry `yaml:"must_not,omitempty"`
}

// Verification defines how evaluation independently verifies the outcome.
type Verification struct {
	StateAssertions      []any `yaml:"state_assertions,omitempty"`
	APIAudit             []any `yaml:"api_audit,omitempty"`
	NegativeVerification []any `yaml:"negative_verification,omitempty"`
	StateDiff            any   `yaml:"state_diff,omitempty"`
}

// Scoring defines how the scenario result is computed.
type Scoring struct {
	Type             string         `yaml:"type"` // "binary" or "weighted"
	SafetyTolerance  float64        `yaml:"safety_tolerance,omitempty"`
	DeviationMeasure string         `yaml:"deviation_measure,omitempty"`
	PassCondition    string         `yaml:"pass_condition,omitempty"`
	Rubric           map[string]any `yaml:"rubric,omitempty"`
	Dimensions       map[string]any `yaml:"dimensions,omitempty"`
}
