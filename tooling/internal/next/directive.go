// Package next provides the deterministic directive for the current governance step.
//
// It reads _status.md, classifies the unit's current lifecycle state,
// and maps that state to a single directive (TASK, READS, WRITES, COMPLETION).
// The directive is the primary agent-facing instruction — the agent should
// execute it directly rather than extracting routing logic from prose.
package next

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

// Outcome describes one valid completion outcome.
type Outcome struct {
	Value string // the --outcome value
	Desc  string // what it means in human terms
}

// Directive is the deterministic instruction for one governance step.
// The agent executes this directly — no routing decisions needed.
type Directive struct {
	Unit       string
	State      UnitState
	Task       string   // one-line task description
	Reads      []string // files the agent should read
	Writes     []string // files the agent may modify (empty = none)
	Blocked    string   // files the agent must not modify
	Completion string   // exact close command template, {unit} placeholder
	Outcomes   []Outcome
	Status     statusfile.ObjectStatus
}

// GetDirective reads the unit's current governance state and returns
// the deterministic directive for that state.
func GetDirective(repoRoot, unitName string) (*Directive, error) {
	status, err := statusfile.LookupObjectStatus(repoRoot, "unit", unitName)
	if err != nil {
		return nil, fmt.Errorf("lookup unit %q: %w", unitName, err)
	}

	state := ClassifyUnitState(status)
	d, err := directiveForState(unitName, state, status)
	if err != nil {
		return nil, err
	}
	d.Status = status
	return d, nil
}

func directiveForState(unitName string, state UnitState, status statusfile.ObjectStatus) (*Directive, error) {
	layer := status.ActiveLayer
	if layer == "" {
		layer = "stable"
	}
	specPath, err := specpaths.ObjectMainSpecFileRef("unit", layer, unitName)
	if err != nil {
		return nil, fmt.Errorf("resolve spec path for %q (layer=%q): %w", unitName, layer, err)
	}
	mappingPath := "docs/specs/repository_mapping.md"

	switch state {
	case StateUnregistered:
		return unregisteredDirective(unitName, specPath, status)
	case StateStableIdle:
		return &Directive{
			Unit:  unitName,
			State: state,
			Task:  fmt.Sprintf("Fork stable truth for candidate change round — %s", unitName),
			Reads: []string{specPath, mappingPath},
			Completion: "specflowctl command close --command unit_fork --object-type unit --object %s --outcome <forked> --apply",
			Outcomes: []Outcome{
				{Value: "forked", Desc: "candidate branch created from stable truth"},
			},
		}, nil
	case StateStableVerify:
		return stableVerifyDirective(unitName, specPath)
	case StateCandidateCheck:
		if status.NextCommand == "unit_new" {
			return &Directive{
				Unit:  unitName,
				State: state,
				Task:  fmt.Sprintf("Create new candidate truth \u2014 %s", unitName),
				Reads: []string{},
				Completion: "specflowctl command close --command unit_new --object-type unit --object %s --outcome <pass|blocked> --apply",
				Outcomes: []Outcome{
					{Value: "pass", Desc: "candidate truth created"},
					{Value: "blocked", Desc: "issue found, cannot create candidate truth"},
				},
			}, nil
		}
		return candidateCheckDirective(unitName, specPath, mappingPath)
	case StateCandidatePending:
		return candidatePendingDirective(unitName, specPath)
	case StateCandidateVerify:
		return candidateVerifyDirective(unitName, specPath, mappingPath)
	case StateCandidatePromote:
		return &Directive{
			Unit:  unitName,
			State: state,
			Task:  fmt.Sprintf("Promote verified candidate truth to stable — %s", unitName),
			Reads: []string{specPath, mappingPath},
			Completion: "specflowctl command close --command unit_promote --object-type unit --object %s --outcome <promoted> --apply",
			Outcomes: []Outcome{
				{Value: "promoted", Desc: "candidate truth promoted to stable"},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unknown unit state %q for %q", state, unitName)
	}
}

func unregisteredDirective(unitName, specPath string, status statusfile.ObjectStatus) (*Directive, error) {
	next := strings.TrimSpace(status.NextCommand)

	switch next {
	case "unit_init":
		return &Directive{
			Unit:  unitName,
			State: StateUnregistered,
			Task:  fmt.Sprintf("Init existing capability \"%s\" as first stable truth", unitName),
			Reads: []string{specPath},
			Completion: "specflowctl command close --command unit_init --object-type unit --object %s --outcome <pass|blocked> --apply",
			Outcomes: []Outcome{
				{Value: "pass", Desc: "stable truth established"},
				{Value: "blocked", Desc: "issue found, cannot establish stable truth"},
			},
		}, nil
	case "unit_new":
		return &Directive{
			Unit:  unitName,
			State: StateUnregistered,
			Task:  fmt.Sprintf("Create new candidate truth — %s", unitName),
			Reads: []string{},
			Completion: "specflowctl command close --command unit_new --object-type unit --object %s --outcome <pass|blocked> --apply",
			Outcomes: []Outcome{
				{Value: "pass", Desc: "candidate truth created"},
				{Value: "blocked", Desc: "issue found, cannot create candidate truth"},
			},
		}, nil
	default:
		return nil, fmt.Errorf("unregistered unit %q has unexpected Next Command %q", unitName, next)
	}
}

func stableVerifyDirective(unitName, specPath string) (*Directive, error) {
	return &Directive{
		Unit:  unitName,
		State: StateStableVerify,
		Task:  fmt.Sprintf("Verify current implementation matches stable truth — %s", unitName),
		Reads: []string{specPath},
		Completion: "specflowctl command close --command unit_stable_verify --object-type unit --object %s --outcome <aligned|controlled_repair_required|controlled_change_required> --apply",
		Outcomes: []Outcome{
			{Value: "aligned", Desc: "implementation matches stable truth"},
			{Value: "controlled_repair_required", Desc: "repair needed, use --candidate-intent repair"},
			{Value: "controlled_change_required", Desc: "change needed, use --candidate-intent change"},
		},
	}, nil
}

func candidateCheckDirective(unitName, specPath, mappingPath string) (*Directive, error) {
	return &Directive{
		Unit:  unitName,
		State: StateCandidateCheck,
		Task:  fmt.Sprintf("Verify candidate spec is clear and ready for implementation — %s", unitName),
		Reads: []string{specPath, mappingPath},
		Completion: "specflowctl command close --command unit_check --object-type unit --object %s --outcome <pass|blocked|fix_required> --apply",
		Outcomes: []Outcome{
			{Value: "pass", Desc: "spec is clear, proceed to implementation"},
			{Value: "blocked", Desc: "spec has issues, list them"},
			{Value: "fix_required", Desc: "spec needs fixes, re-check after fixing"},
		},
	}, nil
}

func candidatePendingDirective(unitName, specPath string) (*Directive, error) {
	mappingPath := "docs/specs/repository_mapping.md"
	return &Directive{
		Unit:  unitName,
		State: StateCandidatePending,
		Task:  fmt.Sprintf("Implement per candidate spec — %s", unitName),
		Reads: []string{specPath, mappingPath},
		Writes: []string{
			"src/**",
			"tests/**",
			mappingPath,
		},
		Blocked: "docs/specs/units/stable/**, docs/specs/_check_result/**, docs/specs/_check_work/**, docs/specs/_verify_result/**, docs/specs/_stable_verify_result/**, docs/specs/_independent_evaluation/**, docs/specs/_plans/**, docs/specs/_status.md, framework/**",
		Completion: "No close command for implementation. Run `specflowctl next --unit %s` when done.",
		Outcomes:  nil,
	}, nil
}

func candidateVerifyDirective(unitName, specPath, mappingPath string) (*Directive, error) {
	verifyResultPath := fmt.Sprintf("docs/specs/_verify_result/unit/%s.md", unitName)
	return &Directive{
		Unit:  unitName,
		State: StateCandidateVerify,
		Task:  fmt.Sprintf("Verify implementation matches candidate spec — %s", unitName),
		Reads: []string{specPath, mappingPath},
		Writes: []string{verifyResultPath},
		Completion: "specflowctl command close --command unit_verify --object-type unit --object %s --outcome <ready_to_promote|spec_issue|impl_issue> --apply",
		Outcomes: []Outcome{
			{Value: "ready_to_promote", Desc: "implementation matches spec, ready for promotion"},
			{Value: "spec_issue", Desc: "spec needs revision, goes back to unit_check"},
			{Value: "impl_issue", Desc: "implementation needs fixes, re-verify after fixing"},
		},
	}, nil
}

// RenderDirective formats the directive for agent consumption.
func RenderDirective(d *Directive) string {
	var buf strings.Builder

	unit := d.Unit
	completion := strings.ReplaceAll(d.Completion, "%s", unit)

	fmt.Fprintf(&buf, "TASK: %s\n\n", d.Task)

	buf.WriteString("READS:\n")
	if len(d.Reads) == 0 {
		buf.WriteString("  (none)\n")
	} else {
		for _, r := range d.Reads {
			fmt.Fprintf(&buf, "  - %s\n", r)
		}
	}
	buf.WriteString("\n")

	buf.WriteString("WRITES:\n")
	if len(d.Writes) == 0 {
		buf.WriteString("  (none)\n")
	} else {
		for _, w := range d.Writes {
			fmt.Fprintf(&buf, "  - %s\n", w)
		}
	}
	buf.WriteString("\n")

	if d.Blocked != "" {
		fmt.Fprintf(&buf, "BLOCKED: %s\n\n", d.Blocked)
	} else {
		buf.WriteString("BLOCKED: All files not listed in READS or WRITES\n\n")
	}

	buf.WriteString("COMPLETION:\n")
	fmt.Fprintf(&buf, "  %s\n", completion)
	if len(d.Outcomes) > 0 {
		for _, o := range d.Outcomes {
			fmt.Fprintf(&buf, "  - %s: %s\n", o.Value, o.Desc)
		}
	}

	return buf.String()
}

// ExplainDirective returns the directive plus lifecycle file content and routing info.
func ExplainDirective(repoRoot string, d *Directive) (string, error) {
	directive := RenderDirective(d)

	lifecycleFile := stateToLifecycleFile(d.State, d.Status)
	if lifecycleFile == "" {
		return directive + "\n=== EXPLAIN ===\nNo lifecycle file for this state.\n", nil
	}

	// Try both source_repo and installed_project layouts
	candidates := []string{
		filepath.Join(repoRoot, "framework/lifecycle", lifecycleFile+".md"),
		filepath.Join(repoRoot, "specflow/framework/lifecycle", lifecycleFile+".md"),
	}

	var content string
	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err == nil {
			content = string(data)
			break
		}
	}

	var buf strings.Builder
	buf.WriteString(directive)
	buf.WriteString("\n=== EXPLAIN ===\n")
	fmt.Fprintf(&buf, "State: %s\n", d.State)
	fmt.Fprintf(&buf, "Lifecycle file: %s.md\n\n", lifecycleFile)
	if content != "" {
		buf.WriteString("--- lifecycle reference ---\n")
		buf.WriteString(content)
		if !strings.HasSuffix(content, "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString("--- end lifecycle reference ---\n")
	} else {
		buf.WriteString("(lifecycle file not found at any expected location)\n")
	}

	return buf.String(), nil
}

// stateToLifecycleFile maps a state to its lifecycle Context Card filename.
func stateToLifecycleFile(state UnitState, status statusfile.ObjectStatus) string {
	switch state {
	case StateUnregistered:
		return "unit_init_new_fork"
	case StateStableIdle:
		return "unit_init_new_fork"
	case StateStableVerify:
		return "unit_stable_verify"
	case StateCandidateCheck:
		if strings.EqualFold(status.NextCommand, "unit_new") || strings.EqualFold(status.NextCommand, "unit_init") {
			return "unit_init_new_fork"
		}
		return "unit_check"
	case StateCandidatePending:
		return "unit_impl"
	case StateCandidateVerify:
		return "unit_verify"
	case StateCandidatePromote:
		return "unit_promote"
	default:
		return ""
	}
}
