package filevalidation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CandidateFrontmatterResult is returned by ValidateCandidateFrontmatter.
type CandidateFrontmatterResult struct {
	Valid      bool
	Unit       string
	Diagnostic string
}

// ValidateCandidateFrontmatter validates candidate unit frontmatter consistency.
//
// Rules from framework/candidate_intent.md:
//
//   - repair  → source_basis=new_design + evidence_appendix_ref=none + repair_basis present
//   - change  → source_basis=existing_implementation|mixed requires evidence_appendix_ref pointing to existing file
//   - change  → source_basis=new_design|replacement requires evidence_appendix_ref=none
//   - unit_new → candidate_intent not required; source_basis rules still apply
//   - repair_basis is not allowed for change candidates
func ValidateCandidateFrontmatter(repoRoot, unitName string) CandidateFrontmatterResult {
	candidatePath := filepath.Join(repoRoot, fmt.Sprintf("docs/specs/units/candidate/c_unit_%s.md", unitName))

	data, err := os.ReadFile(candidatePath)
	if err != nil {
		return CandidateFrontmatterResult{
			Valid:      false,
			Unit:       unitName,
			Diagnostic: fmt.Sprintf("cannot read candidate spec at %s: %v", candidatePath, err),
		}
	}

	fm := readFrontmatterStringMap(string(data))

	candidateIntent := strings.TrimSpace(fm["candidate_intent"])
	sourceBasis := strings.TrimSpace(fm["source_basis"])
	evidenceAppendixRef := strings.TrimSpace(fm["evidence_appendix_ref"])
	repairBasis := strings.TrimSpace(fm["repair_basis"])

	switch candidateIntent {
	case "repair":
		// repair → source_basis=new_design, evidence_appendix_ref=none, repair_basis present
		if repairBasis == "" || repairBasis == "none" {
			return CandidateFrontmatterResult{
				Valid:      false,
				Unit:       unitName,
				Diagnostic: fmt.Sprintf("candidate_intent=repair requires repair_basis (format: s_unit_%s@<version>)", unitName),
			}
		}
		if sourceBasis != "" && sourceBasis != "new_design" {
			return CandidateFrontmatterResult{
				Valid:      false,
				Unit:       unitName,
				Diagnostic: fmt.Sprintf("candidate_intent=repair requires source_basis=new_design, got %q", sourceBasis),
			}
		}
		if evidenceAppendixRef != "" && evidenceAppendixRef != "none" {
			return CandidateFrontmatterResult{
				Valid:      false,
				Unit:       unitName,
				Diagnostic: fmt.Sprintf("candidate_intent=repair requires evidence_appendix_ref=none, got %q", evidenceAppendixRef),
			}
		}

	case "change", "":
		// change or unit_new (no candidate_intent): apply source_basis rules
		switch sourceBasis {
		case "existing_implementation", "mixed":
			// Requires evidence_appendix_ref pointing to a real file
			if evidenceAppendixRef == "" || evidenceAppendixRef == "none" {
				return CandidateFrontmatterResult{
					Valid:      false,
					Unit:       unitName,
					Diagnostic: fmt.Sprintf("source_basis=%s requires evidence_appendix_ref pointing to a valid evidence appendix file", sourceBasis),
				}
			}
			// Verify the evidence appendix file exists
			appendixPath := filepath.Join(repoRoot, fmt.Sprintf("docs/specs/units/candidate/c_unit_%s_%s.md", unitName, evidenceAppendixRef))
			if _, err := os.Stat(appendixPath); err != nil {
				return CandidateFrontmatterResult{
					Valid:      false,
					Unit:       unitName,
					Diagnostic: fmt.Sprintf("evidence_appendix_ref=%s but appendix file not found at %s", evidenceAppendixRef, appendixPath),
				}
			}
		case "new_design", "replacement":
			// Must have evidence_appendix_ref=none
			if evidenceAppendixRef != "" && evidenceAppendixRef != "none" {
				return CandidateFrontmatterResult{
					Valid:      false,
					Unit:       unitName,
					Diagnostic: fmt.Sprintf("source_basis=%s requires evidence_appendix_ref=none, got %q", sourceBasis, evidenceAppendixRef),
				}
			}
		case "":
			return CandidateFrontmatterResult{
				Valid:      false,
				Unit:       unitName,
				Diagnostic: "source_basis is required",
			}
		default:
			return CandidateFrontmatterResult{
				Valid:      false,
				Unit:       unitName,
				Diagnostic: fmt.Sprintf("unsupported source_basis value %q", sourceBasis),
			}
		}

		// repair_basis must be absent for change candidates
		if candidateIntent == "change" && repairBasis != "" && repairBasis != "none" {
			return CandidateFrontmatterResult{
				Valid:      false,
				Unit:       unitName,
				Diagnostic: fmt.Sprintf("candidate_intent=change must not have repair_basis; got %q", repairBasis),
			}
		}

	default:
		return CandidateFrontmatterResult{
			Valid:      false,
			Unit:       unitName,
			Diagnostic: fmt.Sprintf("unsupported candidate_intent value %q; allowed: change, repair", candidateIntent),
		}
	}

	return CandidateFrontmatterResult{
		Valid:      true,
		Unit:       unitName,
		Diagnostic: "candidate frontmatter is valid",
	}
}

// readFrontmatterStringMap reads YAML-like frontmatter (delimited by ---) from text
// and returns key-value pairs as a flat map. List values are ignored.
func readFrontmatterStringMap(text string) map[string]string {
	result := map[string]string{}
	normalized := strings.ReplaceAll(text, "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return result
	}

	for idx := 1; idx < len(lines); idx++ {
		line := lines[idx]
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			break
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "- ") {
			continue
		}
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		// Skip list entries (value empty signals a list key)
		if value == "" {
			continue
		}
		// Trim surrounding quotes and backticks
		value = strings.Trim(value, "`\"' ")
		result[key] = value
	}
	return result
}
