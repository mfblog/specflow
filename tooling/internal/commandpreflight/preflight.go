package commandpreflight

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/snapshot"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

type Result struct {
	Command                string
	ObjectType             string
	Object                 string
	MayContinue            bool
	FailureLayer           string
	RecommendedNextCommand string
	Diagnostics            []string
	ValidatedProcesses     []Process
}

type Process struct {
	ProcessKind            string
	ProcessFile            string
	Result                 string
	FailureLayer           string
	RecommendedNextCommand string
	FreshnessImpact        string
	EvidenceReuse          string
	Diagnostics            []string
}

func Run(repoRoot, command, objectType, object string) Result {
	result := Result{
		Command:                strings.TrimSpace(command),
		ObjectType:             strings.TrimSpace(objectType),
		Object:                 strings.TrimSpace(object),
		MayContinue:            true,
		FailureLayer:           "none",
		RecommendedNextCommand: "none",
	}

	status, err := statusfile.LookupObjectStatus(repoRoot, result.ObjectType, result.Object)
	if err != nil {
		result.MayContinue = false
		result.FailureLayer = "status_layer"
		result.Diagnostics = append(result.Diagnostics, err.Error())
		result.RecommendedNextCommand = "none"
		return result
	}
	if status.NextCommand != result.Command {
		result.MayContinue = false
		result.FailureLayer = "status_layer"
		result.RecommendedNextCommand = status.NextCommand
		result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("status next command mismatch: actual=%s expected=%s", status.NextCommand, result.Command))
		return result
	}

	processKinds, err := ProcessKinds(result.ObjectType, result.Command)
	if err != nil {
		result.MayContinue = false
		result.FailureLayer = "unsupported_command"
		result.Diagnostics = append(result.Diagnostics, err.Error())
		result.RecommendedNextCommand = status.NextCommand
		return result
	}

	for _, processKind := range processKinds {
		process := validateProcess(repoRoot, result.ObjectType, result.Object, processKind)
		result.ValidatedProcesses = append(result.ValidatedProcesses, process)
		if process.Result == "valid" {
			continue
		}
		result.MayContinue = false
		result.FailureLayer = process.FailureLayer
		result.RecommendedNextCommand = noneIfEmpty(process.RecommendedNextCommand)
		result.Diagnostics = append(result.Diagnostics, process.Diagnostics...)
		break
	}

	// Re-validation check: when unit_verify is entered during the implementation
	// phase (Next Command=unit_verify, Notes=pending_impl), detect whether the
	// candidate spec was modified after the last unit_check pass. If the spec
	// fingerprint changed, re-validation via unit_check is required first.
	if result.MayContinue && result.Command == "unit_verify" && strings.Contains(status.Notes, "pending_impl") {
		specPath := filepath.Join(repoRoot, fmt.Sprintf("docs/specs/units/candidate/c_unit_%s.md", result.Object))
		checkResultPath := filepath.Join(repoRoot, fmt.Sprintf("docs/specs/_check_result/unit/%s.md", result.Object))

		// Read the stored truth_fingerprint from _check_result
		checkResultData, err := os.ReadFile(checkResultPath)
		if err == nil {
			storedFingerprint := extractTruthFingerprint(string(checkResultData))
			specData, err := os.ReadFile(specPath)
			currentFingerprint := ""
			if err == nil {
				currentFingerprint = snapshot.ComputeFileFingerprint(string(specData))
			}
			if storedFingerprint != "" && currentFingerprint != "" && currentFingerprint != storedFingerprint {
				result.MayContinue = false
				result.FailureLayer = "gate_layer"
				result.RecommendedNextCommand = "unit_check"
				result.Diagnostics = append(result.Diagnostics,
					fmt.Sprintf("spec modified after unit_check pass: stored truth_fingerprint=%s, current=%s; re-validation required via unit_check:%s",
						storedFingerprint, currentFingerprint, result.Object))
			}
		}
		// If _check_result is absent during re-validation, treat as spec-not-yet-validated
		// (the standard flow has _check_result; its absence in implementation phase is
		// unexpected but unit_check.md permits proceeding with re-validation).
	}

	return result
}

// extractTruthFingerprint parses truth_fingerprint from a _check_result YAML file.
func extractTruthFingerprint(yaml string) string {
	lines := strings.Split(strings.ReplaceAll(yaml, "\r\n", "\n"), "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "truth_fingerprint:") {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "truth_fingerprint:"))
			value = strings.Trim(value, "\"'` ")
			return value
		}
	}
	return ""
}

func ProcessKinds(objectType, command string) ([]string, error) {
	switch objectType {
	case "unit":
		switch command {
		case "unit_init", "unit_new", "unit_fork", "unit_stable_verify", "unit_check":
			return nil, nil
		case "unit_verify":
			return nil, nil
		case "unit_promote":
			return []string{"verify"}, nil
		default:
			return nil, fmt.Errorf("command %q is not supported for object type %q", command, objectType)
		}
	default:
		return nil, fmt.Errorf("object type %q is not supported", objectType)
	}
}

func validateProcess(repoRoot, objectType, object, processKind string) Process {
	process := Process{
		ProcessKind:            processKind,
		Result:                 "invalid",
		FailureLayer:           "tooling_gap",
		RecommendedNextCommand: "none",
	}
	processFile, err := snapshot.ProcessFilePath(objectType, object, processKind)
	if err == nil {
		process.ProcessFile = processFile
	}

	if processFile != "" {
		processAbs := filepath.Join(repoRoot, filepath.FromSlash(processFile))
		if _, err := os.Stat(processAbs); err != nil {
			if os.IsNotExist(err) {
				process.Diagnostics = append(process.Diagnostics, fmt.Sprintf("missing process file: %s", processFile))
				layer, next := fallbackForMissingOrUnavailableProcess(objectType, processKind)
				process.FailureLayer = layer
				process.RecommendedNextCommand = next
				return process
			}
			process.Diagnostics = append(process.Diagnostics, fmt.Sprintf("stat %s: %v", processFile, err))
			return process
		}
	}

	validation, err := snapshot.ValidateProcessFileForObject(repoRoot, objectType, object, processKind)
	if err != nil {
		process.Diagnostics = append(process.Diagnostics, err.Error())
		return process
	}
	process.ProcessFile = validation.ProcessFile
	if validation.Valid {
		process.Result = "valid"
		process.FailureLayer = "none"
		process.RecommendedNextCommand = "none"
		process.FreshnessImpact = validation.FreshnessImpact
		process.EvidenceReuse = validation.EvidenceReuse
		return process
	}
	process.Diagnostics = append(process.Diagnostics, validation.Mismatches...)
	process.FailureLayer = noneIfEmpty(validation.FailureLayer)
	process.RecommendedNextCommand = noneIfEmpty(validation.NextCommand)
	process.FreshnessImpact = validation.FreshnessImpact
	process.EvidenceReuse = validation.EvidenceReuse
	return process
}

func fallbackForMissingOrUnavailableProcess(objectType, processKind string) (string, string) {
	switch objectType {
	case "unit":
		switch processKind {
		case "check":
			return "gate_layer", "unit_check"
		case "verify":
			return "evidence_layer", "unit_verify"
		case "stable_verify":
			return "evidence_layer", "unit_stable_verify"
		}
	}
	return "tooling_gap", "none"
}

func noneIfEmpty(value string) string {
	if value == "" {
		return "none"
	}
	return value
}
