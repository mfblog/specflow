package reader

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var unitCandidateMainSpecPattern = regexp.MustCompile(`^docs/specs/units/candidate/c_unit_([A-Za-z0-9_]+)\.md$`)

type diffOp struct {
	kind      string
	stable    int
	candidate int
	text      string
}

func ReadAllowedSourceDiff(repoRoot, relPath string) (SourceDiff, error) {
	clean, err := cleanAllowedRelativePath(relPath)
	if err != nil {
		return SourceDiff{}, err
	}
	if !isAllowedSourcePath(clean) {
		return SourceDiff{}, fmt.Errorf("source path is not allowed: %s", clean)
	}
	match := unitCandidateMainSpecPattern.FindStringSubmatch(clean)
	if len(match) != 2 {
		return SourceDiff{
			Available:     false,
			CandidatePath: clean,
			Reason:        "not_unit_candidate_main_spec",
		}, nil
	}

	stablePath := "docs/specs/units/stable/s_unit_" + match[1] + ".md"
	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return SourceDiff{}, err
	}
	candidateContent, err := readRepoText(absRoot, clean)
	if err != nil {
		return SourceDiff{}, err
	}
	stableContent, err := readRepoText(absRoot, stablePath)
	if os.IsNotExist(err) {
		return SourceDiff{
			Available:     false,
			CandidatePath: clean,
			StablePath:    stablePath,
			Reason:        "stable_missing",
		}, nil
	}
	if err != nil {
		return SourceDiff{}, err
	}

	hunks, summary := buildLineDiff(splitDiffLines(stableContent), splitDiffLines(candidateContent))
	return SourceDiff{
		Available:     true,
		CandidatePath: clean,
		StablePath:    stablePath,
		Summary:       summary,
		Hunks:         hunks,
	}, nil
}

func readRepoText(absRoot, relPath string) (string, error) {
	absPath := filepath.Join(absRoot, filepath.FromSlash(relPath))
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", fmt.Errorf("source path escapes repo root: %s", relPath)
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(string(data), "\r\n", "\n"), nil
}

func splitDiffLines(text string) []string {
	text = strings.TrimSuffix(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	if text == "" {
		return []string{}
	}
	return strings.Split(text, "\n")
}

func buildLineDiff(stableLines, candidateLines []string) ([]DiffHunk, DiffSummary) {
	ops := lineDiffOps(stableLines, candidateLines)
	hunks := diffHunksFromOps(ops)
	summary := DiffSummary{Hunks: len(hunks)}
	for _, hunk := range hunks {
		hasInsert := false
		hasDelete := false
		for _, line := range hunk.Lines {
			switch line.Type {
			case "insert":
				summary.Added++
				hasInsert = true
			case "delete":
				summary.Deleted++
				hasDelete = true
			}
		}
		if hasInsert && hasDelete {
			summary.Modified++
		}
	}
	return hunks, summary
}

func lineDiffOps(stableLines, candidateLines []string) []diffOp {
	rows := len(stableLines) + 1
	cols := len(candidateLines) + 1
	lcs := make([][]int, rows)
	for i := range lcs {
		lcs[i] = make([]int, cols)
	}
	for i := len(stableLines) - 1; i >= 0; i-- {
		for j := len(candidateLines) - 1; j >= 0; j-- {
			if stableLines[i] == candidateLines[j] {
				lcs[i][j] = lcs[i+1][j+1] + 1
				continue
			}
			if lcs[i+1][j] >= lcs[i][j+1] {
				lcs[i][j] = lcs[i+1][j]
			} else {
				lcs[i][j] = lcs[i][j+1]
			}
		}
	}

	ops := []diffOp{}
	i, j := 0, 0
	for i < len(stableLines) && j < len(candidateLines) {
		if stableLines[i] == candidateLines[j] {
			ops = append(ops, diffOp{kind: "equal", stable: i + 1, candidate: j + 1, text: stableLines[i]})
			i++
			j++
			continue
		}
		if lcs[i+1][j] >= lcs[i][j+1] {
			ops = append(ops, diffOp{kind: "delete", stable: i + 1, text: stableLines[i]})
			i++
		} else {
			ops = append(ops, diffOp{kind: "insert", candidate: j + 1, text: candidateLines[j]})
			j++
		}
	}
	for i < len(stableLines) {
		ops = append(ops, diffOp{kind: "delete", stable: i + 1, text: stableLines[i]})
		i++
	}
	for j < len(candidateLines) {
		ops = append(ops, diffOp{kind: "insert", candidate: j + 1, text: candidateLines[j]})
		j++
	}
	return ops
}

func diffHunksFromOps(ops []diffOp) []DiffHunk {
	const contextLines = 2
	hunks := []DiffHunk{}
	for index := 0; index < len(ops); {
		if ops[index].kind == "equal" {
			index++
			continue
		}
		start := maxInt(0, index-contextLines)
		end := index
		for end < len(ops) && ops[end].kind != "equal" {
			end++
		}
		end = minInt(len(ops), end+contextLines)
		hunks = append(hunks, makeDiffHunk(ops[start:end]))
		index = end
	}
	return hunks
}

func makeDiffHunk(ops []diffOp) DiffHunk {
	hunk := DiffHunk{StableStart: 1, CandidateStart: 1}
	for _, op := range ops {
		if op.stable > 0 {
			hunk.StableStart = op.stable
			break
		}
	}
	for _, op := range ops {
		if op.candidate > 0 {
			hunk.CandidateStart = op.candidate
			break
		}
	}
	for _, op := range ops {
		hunk.Lines = append(hunk.Lines, DiffLine{
			Type:          op.kind,
			StableLine:    op.stable,
			CandidateLine: op.candidate,
			Text:          op.text,
		})
	}
	return hunk
}

func minInt(left, right int) int {
	if left < right {
		return left
	}
	return right
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
