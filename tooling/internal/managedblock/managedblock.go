package managedblock

import (
	"fmt"
	"strings"
)

const (
	BeginMarker = "<!-- SPECFLOW:BEGIN -->"
	EndMarker   = "<!-- SPECFLOW:END -->"
)

func Extract(content string) (string, error) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	beginIdx := -1
	endIdx := -1
	for idx, line := range lines {
		if line == BeginMarker {
			if beginIdx != -1 {
				return "", fmt.Errorf("managed block begin marker must appear exactly once")
			}
			beginIdx = idx
		}
		if line == EndMarker {
			if endIdx != -1 {
				return "", fmt.Errorf("managed block end marker must appear exactly once")
			}
			endIdx = idx
		}
	}
	if beginIdx == -1 || endIdx == -1 || beginIdx >= endIdx {
		return "", fmt.Errorf("managed block markers are missing or out of order")
	}
	return strings.Join(lines[beginIdx:endIdx+1], "\n"), nil
}

func Replace(content, replacement string) (string, error) {
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	beginIdx := -1
	endIdx := -1
	for idx, line := range lines {
		if line == BeginMarker {
			if beginIdx != -1 {
				return "", fmt.Errorf("managed block begin marker must appear exactly once")
			}
			beginIdx = idx
		}
		if line == EndMarker {
			if endIdx != -1 {
				return "", fmt.Errorf("managed block end marker must appear exactly once")
			}
			endIdx = idx
		}
	}
	if beginIdx == -1 || endIdx == -1 || beginIdx >= endIdx {
		return "", fmt.Errorf("managed block markers are missing or out of order")
	}
	replacementLines := strings.Split(strings.ReplaceAll(replacement, "\r\n", "\n"), "\n")
	updated := append([]string{}, lines[:beginIdx]...)
	updated = append(updated, replacementLines...)
	updated = append(updated, lines[endIdx+1:]...)
	return strings.Join(updated, "\n"), nil
}
