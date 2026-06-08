// Package context collects and assembles specFlow context packs for lifecycle commands.
//
// specFlow agents need specific sets of files to perform lifecycle commands
// (unit_impl, unit_verify, etc.). Instead of chain-reading entry_routing.md →
// Context Card → file list, this package lets tooling mechanically collect
// the required files from the repository and produce a structured pack.

package context

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// FileItem represents one file in a context pack.
type FileItem struct {
	Path      string // relative path from repo root
	Essential bool   // true → content is inlined; false → reference link only
	Exists    bool   // whether the file exists on disk
	Content   string // full file content (populated when Essential && Exists)
	LineCount int    // number of lines in the file
}

// Pack is the assembled context for one agent command.
type Pack struct {
	Flow      string     // e.g. "lifecycle"
	Command   string     // e.g. "unit_impl"
	Object    string     // e.g. "auth"
	Files     []FileItem // collected files, sorted by Essential desc then path asc
}

// Collector knows how to collect a context pack for one specific command.
type Collector interface {
	Flow() string
	Command() string
	Collect(repoRoot string, object string) (*Pack, error)
}

// InputRule describes one input file to collect.
type InputRule struct {
	// PathTemplate is a file path template with optional {object} placeholder.
	// When empty, Resolve is used instead.
	PathTemplate string

	// Optional means the file may not exist — skip silently instead of error.
	Optional bool

	// Resolve is an optional function that returns additional file items.
	// When set, PathTemplate is ignored.
	Resolve func(repoRoot string, object string) ([]FileItem, error)

	// Essential controls the Essential flag on the resulting FileItem.
	Essential bool
}

// --- helpers ---

// ResolveFileItem resolves one PathTemplate into a FileItem.
func ResolveFileItem(repoRoot, pathTemplate, object string, essential, optional bool) FileItem {
	path := strings.ReplaceAll(pathTemplate, "{object}", object)
	path = filepath.ToSlash(path)
	absPath := filepath.Join(repoRoot, filepath.FromSlash(path))
	info, err := os.Stat(absPath)
	if err != nil {
		return FileItem{
			Path:      path,
			Essential: essential,
			Exists:    false,
		}
	}
	if info.IsDir() {
		return FileItem{
			Path:      path,
			Essential: essential,
			Exists:    false,
		}
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return FileItem{
			Path:      path,
			Essential: essential,
			Exists:    false,
		}
	}
	content := string(data)
	lineCount := 0
	if content != "" {
		lineCount = strings.Count(content, "\n")
		if !strings.HasSuffix(content, "\n") {
			lineCount++
		}
	}
	return FileItem{
		Path:      path,
		Essential: essential,
		Exists:    true,
		Content:   content,
		LineCount: lineCount,
	}
}

// ResolveInputRules converts a list of InputRules into FileItems.
func ResolveInputRules(repoRoot, object string, rules []InputRule) []FileItem {
	var items []FileItem
	for _, rule := range rules {
		if rule.Resolve != nil {
			resolved, err := rule.Resolve(repoRoot, object)
			if err != nil {
				continue
			}
			items = append(items, resolved...)
			continue
		}
		item := ResolveFileItem(repoRoot, rule.PathTemplate, object, rule.Essential, rule.Optional)
		if !item.Exists && rule.Optional {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Essential != items[j].Essential {
			return items[i].Essential && !items[j].Essential
		}
		return items[i].Path < items[j].Path
	})
	return items
}

// NewPack creates a Pack from resolved items.
func NewPack(flow, command, object string, items []FileItem) *Pack {
	return &Pack{
		Flow:    flow,
		Command: command,
		Object:  object,
		Files:   items,
	}
}

// Inventory returns a summary string of the pack contents.
func (p *Pack) Inventory() string {
	total := len(p.Files)
	essential := 0
	reference := 0
	missing := 0
	for _, f := range p.Files {
		if f.Essential {
			essential++
		} else {
			reference++
		}
		if !f.Exists {
			missing++
		}
	}
	return fmt.Sprintf("Total: %d files (%d inlined, %d referenced, %d missing)", total, essential, reference, missing)
}

// CollectByRules is a convenience helper that builds a pack from input rules.
func CollectByRules(flow, command string, repoRoot, object string, essential, reference []InputRule) (*Pack, error) {
	var all []InputRule
	all = append(all, essential...)
	all = append(all, reference...)
	items := ResolveInputRules(repoRoot, object, all)
	return NewPack(flow, command, object, items), nil
}
