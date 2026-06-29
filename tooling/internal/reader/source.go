package reader

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type SourceFile struct {
	Path      string `json:"path"`
	Content   string `json:"content"`
	LineCount int    `json:"line_count"`
}

func ReadAllowedSource(repoRoot, relPath string) (SourceFile, error) {
	clean, err := cleanAllowedRelativePath(relPath)
	if err != nil {
		return SourceFile{}, err
	}
	if !isAllowedSourcePath(clean) {
		return SourceFile{}, fmt.Errorf("source path is not allowed: %s", clean)
	}

	absRoot, err := filepath.Abs(repoRoot)
	if err != nil {
		return SourceFile{}, err
	}
	absPath := filepath.Join(absRoot, filepath.FromSlash(clean))
	rel, err := filepath.Rel(absRoot, absPath)
	if err != nil {
		return SourceFile{}, err
	}
	if strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return SourceFile{}, fmt.Errorf("source path escapes repo root: %s", clean)
	}

	data, err := os.ReadFile(absPath)
	if err != nil {
		return SourceFile{}, err
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	lineCount := 0
	if text != "" {
		lineCount = strings.Count(text, "\n")
		if !strings.HasSuffix(text, "\n") {
			lineCount++
		}
	}
	return SourceFile{
		Path:      filepath.ToSlash(rel),
		Content:   text,
		LineCount: lineCount,
	}, nil
}

func cleanAllowedRelativePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return "", fmt.Errorf("source path is required")
	}
	path = strings.ReplaceAll(path, "\\", "/")
	if strings.HasPrefix(path, "/") {
		return "", fmt.Errorf("source path must be relative")
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	if clean == "." || strings.HasPrefix(clean, "../") || clean == ".." {
		return "", fmt.Errorf("source path escapes repo root")
	}
	return clean, nil
}

func isAllowedSourcePath(path string) bool {
	for _, prefix := range []string{
		"docs/specs/",
	} {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
