package relationgraph

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

type Result struct {
	RelationResult    string             `json:"relation_result"`
	ReadyCandidates   []string           `json:"ready_candidates"`
	CandidateOrder    []string           `json:"candidate_order"`
	BlockedCandidates []BlockedCandidate `json:"blocked_candidates"`
	CandidateCycles   []CandidateCycle   `json:"candidate_cycles"`
	ReferenceEdges    []ReferenceEdge    `json:"reference_edges"`
	Diagnostics       []string           `json:"diagnostics"`
}

type BlockedCandidate struct {
	Object    string      `json:"object"`
	BlockedBy []string    `json:"blocked_by"`
	Sources   []SourceRef `json:"sources"`
}

type CandidateCycle struct {
	Objects []string    `json:"objects"`
	Sources []SourceRef `json:"sources"`
}

type ReferenceEdge struct {
	FromObject string    `json:"from_object"`
	ToKind     string    `json:"to_kind"`
	ToObject   string    `json:"to_object"`
	Ref        string    `json:"ref"`
	Blocking   bool      `json:"blocking"`
	SourceKind string    `json:"source_kind"`
	Source     SourceRef `json:"source"`
}

type SourceRef struct {
	Path  string `json:"path"`
	Label string `json:"label,omitempty"`
}

type PreflightResult struct {
	RelationResult  string           `json:"relation_result"`
	Object          string           `json:"object"`
	MayContinue     bool             `json:"may_continue"`
	ReadyCandidates []string         `json:"ready_candidates"`
	BlockedBy       []string         `json:"blocked_by"`
	Sources         []SourceRef      `json:"sources"`
	CandidateCycles []CandidateCycle `json:"candidate_cycles"`
	Diagnostics     []string         `json:"diagnostics"`
}

type document struct {
	RelPath     string
	Scalars     map[string]string
	Lists       map[string][]string
	Body        string
	SourceKind  string
	OwnerObject string
}

type candidateUnit struct {
	Object     string
	FileRef    string
	VersionRef string
}

type targetRef struct {
	Kind   string
	Object string
	Ref    string
}

var (
	markdownLinkPattern     = regexp.MustCompile(`\[[^\]]+\]\(([^)]+\.md(?:#[^)]+)?)\)`)
	candidateVersionPattern = regexp.MustCompile("`?(c_unit_[A-Za-z0-9_]+@[0-9]+\\.[0-9]+\\.[0-9]+|c_[gb]_rule_[A-Za-z0-9_]+@[0-9]+\\.[0-9]+\\.[0-9]+)`?")
	stableUnitRefPattern    = regexp.MustCompile("^s_unit_([A-Za-z0-9_]+)@[0-9]+\\.[0-9]+\\.[0-9]+$")
)

func Build(repoRoot string) Result {
	repoRoot, _ = filepath.Abs(repoRoot)
	result := Result{
		RelationResult: "pass",
	}

	statuses, err := statusfile.LoadObjectStatuses(repoRoot)
	if err != nil {
		result.RelationResult = "error"
		result.Diagnostics = append(result.Diagnostics, err.Error())
		return normalize(result)
	}

	candidates := collectCandidateUnits(statuses)
	currentCandidateSet := map[string]bool{}
	for _, candidate := range candidates {
		currentCandidateSet[candidate.Object] = true
	}

	for _, candidate := range candidates {
		mainDoc, ok := readDocument(repoRoot, candidate.FileRef, "main", candidate.Object, &result)
		if !ok {
			continue
		}
		scanDocumentForCandidateRefs(mainDoc, currentCandidateSet, &result)

		for _, doc := range referencedAppendixDocuments(repoRoot, mainDoc, candidate.Object, &result) {
			scanDocumentForCandidateRefs(doc, currentCandidateSet, &result)
		}
	}

	stableCycleDiagnostics(repoRoot, statuses, &result)
	blocked := blockedCandidateMap(result.ReferenceEdges, currentCandidateSet)
	cycles := candidateCycles(result.ReferenceEdges, currentCandidateSet)
	result.CandidateCycles = cycles
	if len(cycles) > 0 {
		result.RelationResult = "fail"
		for _, cycle := range cycles {
			for _, object := range cycle.Objects {
				blocked[object] = appendUniqueString(blocked[object], cycle.Objects...)
			}
		}
	}

	result.ReadyCandidates = readyCandidates(candidates, blocked)
	result.CandidateOrder = candidateOrder(candidates, result.ReferenceEdges, blocked)
	result.BlockedCandidates = blockedCandidates(blocked, result.ReferenceEdges)
	sortReferenceEdges(result.ReferenceEdges)
	return normalize(result)
}

func CandidatePreflight(repoRoot, object string) PreflightResult {
	object = strings.TrimSpace(object)
	graph := Build(repoRoot)
	result := PreflightResult{
		RelationResult:  "pass",
		Object:          object,
		MayContinue:     true,
		ReadyCandidates: append([]string(nil), graph.ReadyCandidates...),
		Sources:         []SourceRef{},
		CandidateCycles: append([]CandidateCycle(nil), graph.CandidateCycles...),
		Diagnostics:     []string{},
	}
	if graph.RelationResult == "error" {
		result.RelationResult = "fail"
		result.MayContinue = false
		result.Diagnostics = append(result.Diagnostics, graph.Diagnostics...)
		return result
	}
	if object == "" {
		result.RelationResult = "fail"
		result.MayContinue = false
		result.Diagnostics = append(result.Diagnostics, "object is required")
		return result
	}
	if containsString(graph.ReadyCandidates, object) {
		return result
	}
	for _, blocked := range graph.BlockedCandidates {
		if blocked.Object == object {
			result.RelationResult = "fail"
			result.MayContinue = false
			result.BlockedBy = append([]string(nil), blocked.BlockedBy...)
			result.Sources = append([]SourceRef(nil), blocked.Sources...)
			result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("%s is blocked by candidate relation dependencies", object))
			return result
		}
	}
	result.RelationResult = "fail"
	result.MayContinue = false
	result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("%s is not a current candidate unit", object))
	return result
}

func collectCandidateUnits(statuses []statusfile.ObjectStatus) []candidateUnit {
	units := []candidateUnit{}
	for _, status := range statuses {
		if status.ObjectType != "unit" || status.Candidate != "yes" || status.ActiveLayer != "candidate" {
			continue
		}
		object := strings.TrimSpace(status.Object)
		if object == "" {
			continue
		}
		units = append(units, candidateUnit{
			Object:  object,
			FileRef: fmt.Sprintf("docs/specs/units/candidate/c_unit_%s.md", object),
		})
	}
	sort.Slice(units, func(i, j int) bool { return units[i].Object < units[j].Object })
	return units
}

func readDocument(repoRoot, relPath, sourceKind, owner string, result *Result) (document, bool) {
	content, err := os.ReadFile(filepath.Join(repoRoot, filepath.FromSlash(relPath)))
	if err != nil {
		result.Diagnostics = append(result.Diagnostics, fmt.Sprintf("read %s: %v", relPath, err))
		return document{}, false
	}
	scalars, lists, body := parseFrontmatterAndBody(string(content))
	return document{
		RelPath:     filepath.ToSlash(relPath),
		Scalars:     scalars,
		Lists:       lists,
		Body:        body,
		SourceKind:  sourceKind,
		OwnerObject: owner,
	}, true
}

func referencedAppendixDocuments(repoRoot string, mainDoc document, owner string, result *Result) []document {
	refs := []SourceRef{}
	if evidenceRef := strings.TrimSpace(mainDoc.Scalars["evidence_appendix_ref"]); evidenceRef != "" && evidenceRef != "none" {
		if resolved, ok := resolveDocRef(mainDoc.RelPath, evidenceRef); ok && isAppendixPath(resolved.Path) {
			resolved.Label = "evidence"
			refs = appendSourceUnique(refs, resolved)
		}
	}
	for _, match := range markdownLinkPattern.FindAllStringSubmatch(mainDoc.Body, -1) {
		if len(match) != 2 {
			continue
		}
		resolved, ok := resolveDocRef(mainDoc.RelPath, match[1])
		if !ok || !isAppendixPath(resolved.Path) {
			continue
		}
		if isEvidencePath(resolved.Path) {
			resolved.Label = "evidence"
		} else {
			resolved.Label = "appendix"
		}
		refs = appendSourceUnique(refs, resolved)
	}
	docs := []document{}
	for _, ref := range refs {
		sourceKind := "appendix"
		if ref.Label == "evidence" || isEvidencePath(ref.Path) {
			sourceKind = "evidence"
		}
		doc, ok := readDocument(repoRoot, ref.Path, sourceKind, owner, result)
		if ok {
			docs = append(docs, doc)
		}
	}
	return docs
}

func scanDocumentForCandidateRefs(doc document, currentCandidates map[string]bool, result *Result) {
	if doc.SourceKind == "main" {
		for _, ref := range frontmatterRefs(doc, "unit_refs") {
			if target, ok := parseCandidateVersionRef(ref); ok {
				addReferenceEdge(doc, target, currentCandidates, result)
			}
		}
		for _, ref := range frontmatterRefs(doc, "rule_refs") {
			if target, ok := parseCandidateVersionRef(ref); ok {
				addReferenceEdge(doc, target, currentCandidates, result)
			}
		}
	}

	for _, match := range markdownLinkPattern.FindAllStringSubmatch(doc.Body, -1) {
		if len(match) != 2 {
			continue
		}
		if resolved, ok := resolveDocRef(doc.RelPath, match[1]); ok {
			if target, ok := targetFromDocPath(resolved.Path); ok {
				target.Ref = resolved.Path
				addReferenceEdge(doc, target, currentCandidates, result)
			}
		}
	}
	for _, match := range candidateVersionPattern.FindAllStringSubmatch(doc.Body, -1) {
		if len(match) != 2 {
			continue
		}
		if target, ok := parseCandidateVersionRef(match[1]); ok {
			addReferenceEdge(doc, target, currentCandidates, result)
		}
	}
}

func addReferenceEdge(doc document, target targetRef, currentCandidates map[string]bool, result *Result) {
	if target.Kind == "" || target.Object == "" {
		return
	}
	if target.Kind == "unit" && target.Object == doc.OwnerObject {
		return
	}
	blocking := doc.SourceKind != "evidence"
	if target.Kind == "unit" && !currentCandidates[target.Object] {
		blocking = doc.SourceKind != "evidence"
	}
	edge := ReferenceEdge{
		FromObject: doc.OwnerObject,
		ToKind:     target.Kind,
		ToObject:   target.Object,
		Ref:        target.Ref,
		Blocking:   blocking,
		SourceKind: doc.SourceKind,
		Source: SourceRef{
			Path:  doc.RelPath,
			Label: doc.SourceKind,
		},
	}
	result.ReferenceEdges = append(result.ReferenceEdges, edge)
}

func parseFrontmatterAndBody(content string) (map[string]string, map[string][]string, string) {
	scalars := map[string]string{}
	lists := map[string][]string{}
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return scalars, lists, content
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		return scalars, lists, content
	}
	currentList := ""
	for _, line := range lines[1:end] {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") && currentList != "" {
			lists[currentList] = append(lists[currentList], trimRef(strings.TrimPrefix(trimmed, "- ")))
			continue
		}
		currentList = ""
		key, value, ok := strings.Cut(trimmed, ":")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if value == "" {
			currentList = key
			continue
		}
		scalars[key] = trimRef(value)
	}
	return scalars, lists, strings.Join(lines[end+1:], "\n")
}

func frontmatterRefs(doc document, key string) []string {
	values := []string{}
	if scalar := strings.TrimSpace(doc.Scalars[key]); scalar != "" && scalar != "none" {
		for _, item := range strings.Split(scalar, ",") {
			item = trimRef(item)
			if item != "" && item != "none" {
				values = append(values, item)
			}
		}
	}
	for _, item := range doc.Lists[key] {
		item = trimRef(item)
		if item != "" && item != "none" {
			values = append(values, item)
		}
	}
	return values
}

func resolveDocRef(fromPath, rawRef string) (SourceRef, bool) {
	rawRef = strings.TrimSpace(rawRef)
	rawRef = strings.Trim(rawRef, "<>")
	if rawRef == "" || strings.Contains(rawRef, "://") || strings.HasPrefix(rawRef, "#") {
		return SourceRef{}, false
	}
	if before, _, ok := strings.Cut(rawRef, "#"); ok {
		rawRef = before
	}
	rawRef = strings.ReplaceAll(rawRef, "\\", "/")
	refPath := rawRef
	if !strings.HasPrefix(refPath, "docs/") {
		refPath = path.Join(path.Dir(fromPath), refPath)
	}
	refPath = path.Clean(refPath)
	if !strings.HasPrefix(refPath, "docs/specs/") {
		return SourceRef{}, false
	}
	return SourceRef{Path: refPath}, true
}

func targetFromDocPath(refPath string) (targetRef, bool) {
	base := path.Base(refPath)
	switch {
	case strings.HasPrefix(refPath, "docs/specs/units/candidate/") && strings.HasPrefix(base, "c_unit_") && strings.HasSuffix(base, ".md") && !strings.Contains(refPath, "/appendix/"):
		return targetRef{Kind: "unit", Object: strings.TrimSuffix(strings.TrimPrefix(base, "c_unit_"), ".md")}, true
	case strings.HasPrefix(refPath, "docs/specs/rules/candidate/") && strings.HasPrefix(base, "c_") && strings.HasSuffix(base, ".md"):
		return targetRef{Kind: "rule", Object: strings.TrimPrefix(strings.TrimSuffix(base, ".md"), "c_")}, true
	default:
		return targetRef{}, false
	}
}

func parseCandidateVersionRef(ref string) (targetRef, bool) {
	ref = trimRef(ref)
	prefix, _, ok := strings.Cut(ref, "@")
	if !ok {
		return targetRef{}, false
	}
	switch {
	case strings.HasPrefix(prefix, "c_unit_"):
		return targetRef{Kind: "unit", Object: strings.TrimPrefix(prefix, "c_unit_"), Ref: ref}, true
	case strings.HasPrefix(prefix, "c_b_rule_") || strings.HasPrefix(prefix, "c_g_rule_"):
		return targetRef{Kind: "rule", Object: strings.TrimPrefix(prefix, "c_"), Ref: ref}, true
	default:
		return targetRef{}, false
	}
}

func blockedCandidateMap(edges []ReferenceEdge, currentCandidates map[string]bool) map[string][]string {
	blocked := map[string][]string{}
	for _, edge := range edges {
		if !edge.Blocking {
			continue
		}
		if edge.ToKind == "unit" {
			blocked[edge.FromObject] = appendUniqueString(blocked[edge.FromObject], "unit:"+edge.ToObject)
			continue
		}
		if edge.ToKind == "rule" {
			blocked[edge.FromObject] = appendUniqueString(blocked[edge.FromObject], "rule:"+edge.ToObject)
		}
	}
	for object, deps := range blocked {
		if !currentCandidates[object] {
			delete(blocked, object)
			continue
		}
		sort.Strings(deps)
		blocked[object] = deps
	}
	return blocked
}

func readyCandidates(candidates []candidateUnit, blocked map[string][]string) []string {
	ready := []string{}
	for _, candidate := range candidates {
		if len(blocked[candidate.Object]) == 0 {
			ready = append(ready, candidate.Object)
		}
	}
	sort.Strings(ready)
	return ready
}

func candidateOrder(candidates []candidateUnit, edges []ReferenceEdge, blocked map[string][]string) []string {
	candidateSet := map[string]bool{}
	remaining := map[string]bool{}
	for _, candidate := range candidates {
		candidateSet[candidate.Object] = true
		remaining[candidate.Object] = true
	}
	deps := map[string][]string{}
	for _, edge := range edges {
		if !edge.Blocking || edge.ToKind != "unit" || !candidateSet[edge.ToObject] {
			continue
		}
		deps[edge.FromObject] = appendUniqueString(deps[edge.FromObject], edge.ToObject)
	}
	order := []string{}
	for len(remaining) > 0 {
		layer := []string{}
		for object := range remaining {
			if hasExternalBlocker(blocked[object]) {
				continue
			}
			if dependenciesCleared(deps[object], remaining) {
				layer = append(layer, object)
			}
		}
		if len(layer) == 0 {
			break
		}
		sort.Strings(layer)
		for _, object := range layer {
			order = append(order, object)
			delete(remaining, object)
		}
	}
	return order
}

func hasExternalBlocker(blockers []string) bool {
	for _, blocker := range blockers {
		if strings.HasPrefix(blocker, "rule:") {
			return true
		}
	}
	return false
}

func dependenciesCleared(deps []string, remaining map[string]bool) bool {
	for _, dep := range deps {
		if remaining[dep] {
			return false
		}
	}
	return true
}

func blockedCandidates(blocked map[string][]string, edges []ReferenceEdge) []BlockedCandidate {
	result := make([]BlockedCandidate, 0, len(blocked))
	for object, blockers := range blocked {
		sources := []SourceRef{}
		for _, edge := range edges {
			if edge.FromObject != object || !edge.Blocking {
				continue
			}
			sources = appendSourceUnique(sources, edge.Source)
		}
		sort.Slice(sources, func(i, j int) bool { return sources[i].Path < sources[j].Path })
		sort.Strings(blockers)
		result = append(result, BlockedCandidate{
			Object:    object,
			BlockedBy: blockers,
			Sources:   sources,
		})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Object < result[j].Object })
	return result
}

func candidateCycles(edges []ReferenceEdge, currentCandidates map[string]bool) []CandidateCycle {
	graph := map[string][]string{}
	for _, edge := range edges {
		if !edge.Blocking || edge.ToKind != "unit" {
			continue
		}
		if !currentCandidates[edge.FromObject] || !currentCandidates[edge.ToObject] {
			continue
		}
		graph[edge.FromObject] = appendUniqueString(graph[edge.FromObject], edge.ToObject)
	}
	cycles := [][]string{}
	seenCycles := map[string]bool{}
	var visit func(start, current string, path []string)
	visit = func(start, current string, chain []string) {
		for _, next := range graph[current] {
			if next == start {
				cycle := append(append([]string(nil), chain...), start)
				key := canonicalCycleKey(cycle[:len(cycle)-1])
				if !seenCycles[key] {
					seenCycles[key] = true
					cycles = append(cycles, cycle)
				}
				continue
			}
			if containsString(chain, next) {
				continue
			}
			visit(start, next, append(append([]string(nil), chain...), next))
		}
	}
	nodes := make([]string, 0, len(currentCandidates))
	for node := range currentCandidates {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)
	for _, node := range nodes {
		visit(node, node, []string{node})
	}
	result := []CandidateCycle{}
	for _, cycle := range cycles {
		sources := []SourceRef{}
		for idx := 0; idx < len(cycle)-1; idx++ {
			from := cycle[idx]
			to := cycle[idx+1]
			for _, edge := range edges {
				if edge.FromObject == from && edge.ToKind == "unit" && edge.ToObject == to && edge.Blocking {
					sources = appendSourceUnique(sources, edge.Source)
				}
			}
		}
		result = append(result, CandidateCycle{Objects: cycle, Sources: sources})
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.Join(result[i].Objects, "->") < strings.Join(result[j].Objects, "->")
	})
	return result
}

func canonicalCycleKey(cycle []string) string {
	if len(cycle) == 0 {
		return ""
	}
	best := append([]string(nil), cycle...)
	for i := 1; i < len(cycle); i++ {
		rotated := append(append([]string(nil), cycle[i:]...), cycle[:i]...)
		if strings.Join(rotated, "\x00") < strings.Join(best, "\x00") {
			best = rotated
		}
	}
	return strings.Join(best, "->")
}

func stableCycleDiagnostics(repoRoot string, statuses []statusfile.ObjectStatus, result *Result) {
	graph := map[string][]string{}
	for _, status := range statuses {
		if status.ObjectType != "unit" {
			continue
		}
		layer := status.ActiveLayer
		if layer != "candidate" && layer != "stable" {
			continue
		}
		fileRef := fmt.Sprintf("docs/specs/units/%s/%s_unit_%s.md", layer, layerPrefix(layer), status.Object)
		doc, ok := readDocument(repoRoot, fileRef, "main", status.Object, result)
		if !ok {
			continue
		}
		for _, ref := range frontmatterRefs(doc, "unit_refs") {
			matches := stableUnitRefPattern.FindStringSubmatch(trimRef(ref))
			if len(matches) == 2 {
				graph[status.Object] = appendUniqueString(graph[status.Object], matches[1])
			}
		}
	}
	for _, component := range stronglyConnectedComponents(graph) {
		result.Diagnostics = append(result.Diagnostics, "stable_reference_cycle: "+strings.Join(component, " <-> "))
	}
}

func stronglyConnectedComponents(graph map[string][]string) [][]string {
	nodeSet := map[string]bool{}
	for node, deps := range graph {
		nodeSet[node] = true
		for _, dep := range deps {
			nodeSet[dep] = true
		}
	}
	nodes := make([]string, 0, len(nodeSet))
	for node := range nodeSet {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)

	index := 0
	indexes := map[string]int{}
	lowlinks := map[string]int{}
	onStack := map[string]bool{}
	stack := []string{}
	components := [][]string{}

	var strongConnect func(node string)
	strongConnect = func(node string) {
		indexes[node] = index
		lowlinks[node] = index
		index++
		stack = append(stack, node)
		onStack[node] = true

		deps := append([]string(nil), graph[node]...)
		sort.Strings(deps)
		for _, dep := range deps {
			if _, seen := indexes[dep]; !seen {
				strongConnect(dep)
				if lowlinks[dep] < lowlinks[node] {
					lowlinks[node] = lowlinks[dep]
				}
				continue
			}
			if onStack[dep] && indexes[dep] < lowlinks[node] {
				lowlinks[node] = indexes[dep]
			}
		}

		if lowlinks[node] != indexes[node] {
			return
		}
		component := []string{}
		for {
			last := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			onStack[last] = false
			component = append(component, last)
			if last == node {
				break
			}
		}
		sort.Strings(component)
		if len(component) > 1 || containsString(graph[node], node) {
			components = append(components, component)
		}
	}

	for _, node := range nodes {
		if _, seen := indexes[node]; !seen {
			strongConnect(node)
		}
	}
	sort.Slice(components, func(i, j int) bool {
		return strings.Join(components[i], "\x00") < strings.Join(components[j], "\x00")
	})
	return components
}

func layerPrefix(layer string) string {
	if layer == "stable" {
		return "s"
	}
	return "c"
}

func sortReferenceEdges(edges []ReferenceEdge) {
	sort.Slice(edges, func(i, j int) bool {
		left := edges[i]
		right := edges[j]
		return strings.Join([]string{left.FromObject, left.ToKind, left.ToObject, left.Source.Path, left.Ref}, "\x00") <
			strings.Join([]string{right.FromObject, right.ToKind, right.ToObject, right.Source.Path, right.Ref}, "\x00")
	})
}

func normalize(result Result) Result {
	if result.ReadyCandidates == nil {
		result.ReadyCandidates = []string{}
	}
	if result.CandidateOrder == nil {
		result.CandidateOrder = []string{}
	}
	if result.BlockedCandidates == nil {
		result.BlockedCandidates = []BlockedCandidate{}
	}
	if result.CandidateCycles == nil {
		result.CandidateCycles = []CandidateCycle{}
	}
	if result.ReferenceEdges == nil {
		result.ReferenceEdges = []ReferenceEdge{}
	}
	if result.Diagnostics == nil {
		result.Diagnostics = []string{}
	}
	return result
}

func appendUniqueString(items []string, values ...string) []string {
	seen := map[string]bool{}
	for _, item := range items {
		seen[item] = true
	}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		items = append(items, value)
		seen[value] = true
	}
	return items
}

func appendSourceUnique(items []SourceRef, values ...SourceRef) []SourceRef {
	seen := map[string]bool{}
	for _, item := range items {
		seen[item.Path] = true
	}
	for _, value := range values {
		if value.Path == "" || seen[value.Path] {
			continue
		}
		items = append(items, value)
		seen[value.Path] = true
	}
	return items
}

func containsString(items []string, value string) bool {
	for _, item := range items {
		if item == value {
			return true
		}
	}
	return false
}

func isAppendixPath(relPath string) bool {
	return strings.Contains(relPath, "/appendix/")
}

func isEvidencePath(relPath string) bool {
	return strings.Contains(relPath, "/appendix/") && strings.HasSuffix(path.Base(relPath), "_evidence.md")
}

func trimRef(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, "`")
	value = strings.Trim(value, "\"")
	value = strings.Trim(value, "'")
	return strings.TrimSpace(value)
}
