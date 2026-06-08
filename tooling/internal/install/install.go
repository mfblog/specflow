package install

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/buildrelease"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/managedblock"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/manifest"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specflowlayout"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/toolingfreshness"
)

type InitResult struct {
	Copied  int
	Skipped int
}

type DoctorResult struct {
	Failures []string
	Warnings []string
}

func Init(repoRoot string, force bool) (InitResult, error) {
	layout, err := specflowlayout.Resolve(repoRoot)
	if err != nil {
		return InitResult{}, err
	}
	items, err := manifest.Load(repoRoot)
	if err != nil {
		return InitResult{}, err
	}

	result := InitResult{}
	for _, item := range items {
		sourceRelative := specflowlayout.Relative(layout.ContentRoot, item.SourceRelative)
		source := filepath.Join(repoRoot, filepath.FromSlash(sourceRelative))
		dest := filepath.Join(repoRoot, filepath.FromSlash(item.DestinationRelative))
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return result, fmt.Errorf("mkdir %s: %w", item.DestinationRelative, err)
		}
		if _, err := os.Stat(dest); err == nil {
			if isManagedEntryFile(item.DestinationRelative) {
				if err := syncManagedEntryFile(source, dest); err != nil {
					return result, fmt.Errorf("install managed block %s: %w", item.DestinationRelative, err)
				}
				result.Copied++
				continue
			}
			if !force {
				result.Skipped++
				continue
			}
		}

		if err := copyFile(source, dest); err != nil {
			return result, fmt.Errorf("copy %s: %w", item.DestinationRelative, err)
		}
		result.Copied++
	}

	return result, nil
}

func Doctor(repoRoot string) (DoctorResult, error) {
	layout, err := specflowlayout.Resolve(repoRoot)
	if err != nil {
		return DoctorResult{}, err
	}
	items, err := manifest.Load(repoRoot)
	if err != nil {
		return DoctorResult{}, err
	}

	result := DoctorResult{}
	for _, item := range items {
		expectedRelative := item.DestinationRelative
		if layout.Kind == specflowlayout.SourceRepo {
			expectedRelative = specflowlayout.Relative(layout.ContentRoot, item.SourceRelative)
		}
		expected := filepath.Join(repoRoot, filepath.FromSlash(expectedRelative))
		if _, err := os.Stat(expected); err != nil {
			result.Failures = append(result.Failures, fmt.Sprintf("MISSING %s", expectedRelative))
		}
	}

	if err := checkManagedEntryConsistency(repoRoot, layout, &result); err != nil {
		return result, err
	}
	checkBinary(repoRoot, layout, &result)
	checkReaderWeb(repoRoot, layout, &result)
	return result, nil
}

func isManagedEntryFile(path string) bool {
	switch filepath.ToSlash(path) {
	case "AGENTS.md", "GEMINI.md", "CLAUDE.md":
		return true
	default:
		return false
	}
}

func syncManagedEntryFile(source, dest string) error {
	sourceContent, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	block, err := managedblock.Extract(string(sourceContent))
	if err != nil {
		return err
	}

	destContent, err := os.ReadFile(dest)
	if err != nil {
		return err
	}
	destText := string(destContent)
	hasBegin := strings.Contains(destText, managedblock.BeginMarker)
	hasEnd := strings.Contains(destText, managedblock.EndMarker)

	switch {
	case hasBegin && hasEnd:
		updated, err := managedblock.Replace(destText, block)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, []byte(updated), 0o644)
	case !hasBegin && !hasEnd:
		if strings.HasSuffix(destText, "\r\n") {
			destText = strings.TrimSuffix(destText, "\r\n") + "\n"
		}
		if strings.TrimSpace(destText) == "" {
			destText = block + "\n"
		} else if strings.HasSuffix(destText, "\n") {
			destText = block + "\n\n" + strings.TrimRight(destText, "\n") + "\n"
		} else {
			destText = block + "\n\n" + destText + "\n"
		}
		return os.WriteFile(dest, []byte(destText), 0o644)
	default:
		return fmt.Errorf("managed block markers are incomplete in destination file")
	}
}

func copyFile(source, dest string) error {
	content, err := os.ReadFile(source)
	if err != nil {
		return err
	}
	return os.WriteFile(dest, content, 0o644)
}

func checkManagedEntryConsistency(repoRoot string, layout specflowlayout.Layout, result *DoctorResult) error {
	entryRoot := repoRoot
	if layout.Kind == specflowlayout.SourceRepo {
		entryRoot = filepath.Join(repoRoot, filepath.FromSlash(layout.TemplateRoot))
	}
	agentsPath := filepath.Join(entryRoot, "AGENTS.md")
	if _, err := os.Stat(agentsPath); err != nil {
		return nil
	}

	agentsContent, err := os.ReadFile(agentsPath)
	if err != nil {
		return err
	}
	agentsBlock, err := managedblock.Extract(string(agentsContent))
	if err != nil {
		result.Failures = append(result.Failures, "INVALID managed block in AGENTS.md")
		return nil
	}

	for _, peer := range []string{"GEMINI.md", "CLAUDE.md"} {
		peerPath := filepath.Join(entryRoot, peer)
		if _, err := os.Stat(peerPath); err != nil {
			continue
		}
		peerContent, err := os.ReadFile(peerPath)
		if err != nil {
			return err
		}
		peerBlock, err := managedblock.Extract(string(peerContent))
		if err != nil {
			result.Failures = append(result.Failures, fmt.Sprintf("INVALID managed block in %s", peer))
			continue
		}
		if agentsBlock != peerBlock {
			result.Failures = append(result.Failures, fmt.Sprintf("DIFF managed blocks in AGENTS.md and %s", peer))
		}
	}
	return nil
}

func checkBinary(repoRoot string, layout specflowlayout.Layout, result *DoctorResult) {
	checkOneBinary(repoRoot, specflowlayout.Relative(layout.ToolingRoot, filepath.ToSlash(filepath.Join("bin", buildrelease.CurrentBinaryName()))), result)
	checkOneBinary(repoRoot, specflowlayout.Relative(layout.ToolingRoot, filepath.ToSlash(filepath.Join("bin", buildrelease.CurrentReaderBinaryName()))), result)
}

func checkOneBinary(repoRoot, relPath string, result *DoctorResult) {
	binaryPath := filepath.Join(repoRoot, filepath.FromSlash(relPath))
	if _, err := os.Stat(binaryPath); err != nil {
		result.Failures = append(result.Failures, fmt.Sprintf("MISSING %s", relPath))
		return
	}

	liveFingerprint, _, err := toolingfreshness.LiveFingerprint(repoRoot)
	if err != nil {
		result.Failures = append(result.Failures, fmt.Sprintf("INVALID tooling live fingerprint: %v", err))
		return
	}

	builtFingerprint, err := toolingfreshness.ReadBuildFingerprintFromBinary(binaryPath)
	if err != nil {
		result.Failures = append(result.Failures, fmt.Sprintf("INVALID %s freshness probe failed: %v", relPath, err))
		return
	}
	if strings.TrimSpace(builtFingerprint) == "" {
		result.Failures = append(result.Failures, fmt.Sprintf("INVALID %s missing embedded build fingerprint", relPath))
		return
	}
	if strings.TrimSpace(builtFingerprint) != strings.TrimSpace(liveFingerprint) {
		result.Failures = append(result.Failures, fmt.Sprintf(
			"STALE %s built_fingerprint=%s live_fingerprint=%s",
			relPath,
			shortFingerprint(builtFingerprint),
			shortFingerprint(liveFingerprint),
		))
	}
}

func checkReaderWeb(repoRoot string, layout specflowlayout.Layout, result *DoctorResult) {
	for _, asset := range []string{
		"index.html",
		"styles.css",
		"app.js",
		"cytoscape.min.js",
		"mermaid.min.js",
	} {
		relPath := specflowlayout.Relative(layout.ToolingRoot, filepath.ToSlash(filepath.Join("reader", "web", asset)))
		path := filepath.Join(repoRoot, filepath.FromSlash(relPath))
		info, err := os.Stat(path)
		if err != nil {
			result.Failures = append(result.Failures, fmt.Sprintf("MISSING %s", relPath))
			continue
		}
		if info.IsDir() {
			result.Failures = append(result.Failures, fmt.Sprintf("INVALID %s is a directory", relPath))
		}
	}
}

func shortFingerprint(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 12 {
		return value
	}
	return value[:12]
}
