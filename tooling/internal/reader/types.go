package reader

import "github.com/Bingordinary/SpecFlow/specflow/tooling/internal/relationgraph"

type Snapshot struct {
	Version            int64                `json:"version"`
	GeneratedAt        string               `json:"generated_at"`
	Project            ProjectInfo          `json:"project"`
	Objects            []ObjectView         `json:"objects"`
	Registry           []RegistryItem       `json:"registry"`
	CandidateRelations relationgraph.Result `json:"candidate_relations"`
	Nodes              []GraphNode          `json:"nodes"`
	Edges              []GraphEdge          `json:"edges"`
	Sources            []SourceRef          `json:"sources"`
	Diagnostics        []Diagnostic         `json:"diagnostics"`
}

type ProjectInfo struct {
	RepoRoot         string `json:"repo_root"`
	StatusFile       string `json:"status_file"`
	MappingFile      string `json:"mapping_file"`
	RuleBaselineFile string `json:"rule_baseline_file"`
	UnitCount        int    `json:"unit_count"`
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
	UnitRefs            []string    `json:"unit_refs"`
	BoundObjects        []string    `json:"bound_objects"`
	Sources             []SourceRef `json:"sources"`
}

type RegistryItem struct {
	ID                       string      `json:"id"`
	Kind                     string      `json:"kind"`
	Label                    string      `json:"label"`
	RuleScope                string      `json:"rule_scope,omitempty"`
	RegistrationState        string      `json:"registration_state,omitempty"`
	Result                   string      `json:"result"`
	MappingRegistered        bool        `json:"mapping_registered"`
	StatusRegistered         bool        `json:"status_registered"`
	TruthRegistered          bool        `json:"truth_registered"`
	ImplementationRegistered bool        `json:"implementation_registered"`
	MappingSource            *SourceRef  `json:"mapping_source,omitempty"`
	StatusSource             *SourceRef  `json:"status_source,omitempty"`
	TruthSources             []SourceRef `json:"truth_sources"`
	ImplementationPaths      []SourceRef `json:"implementation_paths"`
	RuleRefs                 []string    `json:"rule_refs"`
	UnitRefs                 []string    `json:"unit_refs"`
	BoundObjects             []string    `json:"bound_objects"`
	Issues                   []string    `json:"issues"`
	Sources                  []SourceRef `json:"sources"`
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

type SourceDiff struct {
	Available     bool        `json:"available"`
	CandidatePath string      `json:"candidate_path,omitempty"`
	StablePath    string      `json:"stable_path,omitempty"`
	Reason        string      `json:"reason,omitempty"`
	Summary       DiffSummary `json:"summary"`
	Hunks         []DiffHunk  `json:"hunks"`
}

type DiffSummary struct {
	Added    int `json:"added"`
	Deleted  int `json:"deleted"`
	Modified int `json:"modified"`
	Hunks    int `json:"hunks"`
}

type DiffHunk struct {
	StableStart    int        `json:"stable_start"`
	CandidateStart int        `json:"candidate_start"`
	Lines          []DiffLine `json:"lines"`
}

type DiffLine struct {
	Type          string `json:"type"`
	StableLine    int    `json:"stable_line,omitempty"`
	CandidateLine int    `json:"candidate_line,omitempty"`
	Text          string `json:"text"`
}
