package reader

type Snapshot struct {
	Version     int64        `json:"version"`
	GeneratedAt string       `json:"generated_at"`
	Project     ProjectInfo  `json:"project"`
	Objects     []ObjectView `json:"objects"`
	Nodes       []GraphNode  `json:"nodes"`
	Edges       []GraphEdge  `json:"edges"`
	Sources     []SourceRef  `json:"sources"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

type ProjectInfo struct {
	RepoRoot         string `json:"repo_root"`
	StatusFile       string `json:"status_file"`
	MappingFile      string `json:"mapping_file"`
	RuleBaselineFile string `json:"rule_baseline_file"`
	UnitCount        int    `json:"unit_count"`
	ScenarioCount    int    `json:"scenario_count"`
	RuleCount        int    `json:"rule_count"`
	TruthFileCount   int    `json:"truth_file_count"`
}

type ObjectView struct {
	ID                  string      `json:"id"`
	Kind                string      `json:"kind"`
	Label               string      `json:"label"`
	Responsibility      string      `json:"responsibility,omitempty"`
	Layer               string      `json:"layer,omitempty"`
	Version             string      `json:"version,omitempty"`
	HumanState          string      `json:"human_state,omitempty"`
	Stable              string      `json:"stable,omitempty"`
	Candidate           string      `json:"candidate,omitempty"`
	NextCommand         string      `json:"next_command,omitempty"`
	NextLabel           string      `json:"next_label,omitempty"`
	NextIntent          string      `json:"next_intent,omitempty"`
	NextIntentLabel     string      `json:"next_intent_label,omitempty"`
	Notes               string      `json:"notes,omitempty"`
	TruthPaths          []SourceRef `json:"truth_paths"`
	ImplementationPaths []SourceRef `json:"implementation_paths"`
	RuleRefs            []string    `json:"rule_refs"`
	BoundObjects        []string    `json:"bound_objects"`
	Sources             []SourceRef `json:"sources"`
}

type GraphNode struct {
	ID     string     `json:"id"`
	Kind   string     `json:"kind"`
	Label  string     `json:"label"`
	Group  string     `json:"group"`
	Source *SourceRef `json:"source,omitempty"`
}

type GraphEdge struct {
	ID     string     `json:"id"`
	From   string     `json:"from"`
	To     string     `json:"to"`
	Kind   string     `json:"kind"`
	Label  string     `json:"label"`
	Source *SourceRef `json:"source,omitempty"`
}

type SourceRef struct {
	Path  string `json:"path"`
	Line  int    `json:"line,omitempty"`
	Label string `json:"label,omitempty"`
}

type Diagnostic struct {
	Severity string     `json:"severity"`
	Message  string     `json:"message"`
	Source   *SourceRef `json:"source,omitempty"`
}
