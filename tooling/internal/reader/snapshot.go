package reader

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/specpaths"
	"github.com/Bingordinary/SpecFlow/specflow/tooling/internal/statusfile"
)

type markdownDoc struct {
	RelPath     string
	Title       string
	Frontmatter frontmatter
	Text        string
}

var sharedRefPattern = regexp.MustCompile("`?([cs]_[gb]_rule_[A-Za-z0-9_]+@[0-9]+\\.[0-9]+\\.[0-9]+)`?")
var markdownLinkPattern = regexp.MustCompile(`\[[^\]]+\]\(([^)]+\.md(?:#[^)]+)?)\)`)
var candidateIntentPattern = regexp.MustCompile(`\bcandidate[_-]intent\s*=?\s*(repair|change)\b`)

func BuildSnapshot(repoRoot string) Snapshot {
	repoRoot, _ = filepath.Abs(repoRoot)
	mapping := loadRepositoryMapping(repoRoot)
	docs, docDiagnostics := loadTruthDocs(repoRoot)
	docByPath := map[string]markdownDoc{}
	for _, doc := range docs {
		docByPath[doc.RelPath] = doc
	}

	snapshot := Snapshot{
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Project: ProjectInfo{
			RepoRoot:         repoRoot,
			StatusFile:       "docs/specs/_status.md",
			MappingFile:      "docs/specs/repository_mapping.md",
			RuleBaselineFile: "docs/specs/rules/stable/s_g_rule_repository_baseline.md",
		},
		Diagnostics: append(mapping.Diagnostics, docDiagnostics...),
	}

	statuses, err := statusfile.LoadObjectStatuses(repoRoot)
	if err != nil {
		snapshot.Diagnostics = append(snapshot.Diagnostics, Diagnostic{
			Severity: "error",
			Message:  "cannot read status table: " + err.Error(),
			Source:   &SourceRef{Path: "docs/specs/_status.md"},
		})
	}

	builder := newGraphBuilder()
	sourceSet := map[string]SourceRef{}
	addSource := func(ref SourceRef) {
		if ref.Path == "" {
			return
		}
		key := ref.Path
		if existing, ok := sourceSet[key]; !ok || existing.Line == 0 {
			sourceSet[key] = ref
		}
	}
	addSource(SourceRef{Path: "docs/specs/_status.md", Label: "Status"})
	addSource(SourceRef{Path: "docs/specs/repository_mapping.md", Label: "Repository Mapping"})
	addSource(SourceRef{Path: "docs/specs/rules/stable/s_g_rule_repository_baseline.md", Label: "Global Rules"})

	for _, doc := range docs {
		addSource(SourceRef{Path: doc.RelPath, Label: doc.Title})
	}

	builder.addNode(GraphNode{ID: "rule:baseline", Kind: "rule", Label: "Global Rules", Group: "rule", Source: ptr(SourceRef{Path: "docs/specs/rules/stable/s_g_rule_repository_baseline.md"})})

	for _, status := range statuses {
		if status.ObjectType == "unit" {
			snapshot.Project.UnitCount++
		}
		if status.ObjectType == "scenario" {
			snapshot.Project.ScenarioCount++
		}
		object := buildObjectFromStatus(status, mapping, docByPath)
		snapshot.Objects = append(snapshot.Objects, object)
		nodeID := status.ObjectType + ":" + status.Object
		builder.addNode(GraphNode{ID: nodeID, Kind: status.ObjectType, Label: status.Object, Group: status.ObjectType, Source: ptr(SourceRef{Path: "docs/specs/_status.md"})})
		for _, truth := range object.TruthPaths {
			addSource(truth)
			fileNode := "file:" + truth.Path
			builder.addNode(GraphNode{ID: fileNode, Kind: "truth_file", Label: compactTruthFileLabel(filepath.Base(truth.Path)), Group: "truth", Source: ptr(SourceRef{Path: truth.Path})})
			builder.addEdge(GraphEdge{ID: nodeID + "->" + fileNode, From: nodeID, To: fileNode, Kind: "described_by", Label: "described by", Source: ptr(truth)})
		}
		for _, impl := range object.ImplementationPaths {
			pathNode := "path:" + impl.Path
			builder.addNode(GraphNode{ID: pathNode, Kind: "implementation_path", Label: impl.Path, Group: "implementation", Source: ptr(impl)})
			builder.addEdge(GraphEdge{ID: nodeID + "->" + pathNode, From: nodeID, To: pathNode, Kind: "owns_path", Label: "owns path", Source: ptr(impl)})
		}
		for _, sharedID := range object.RuleRefs {
			sharedNode := "shared:" + sharedID
			builder.addNode(GraphNode{ID: sharedNode, Kind: "rule", Label: sharedID, Group: "shared"})
			builder.addEdge(GraphEdge{ID: nodeID + "->" + sharedNode, From: nodeID, To: sharedNode, Kind: "uses_shared", Label: "uses shared", Source: firstSource(object.Sources)})
		}
	}

	sharedObjects := buildSharedObjects(mapping, docs)
	snapshot.Project.RuleCount = len(sharedObjects)
	for _, object := range sharedObjects {
		snapshot.Objects = append(snapshot.Objects, object)
		sharedNode := "shared:" + object.ID
		builder.addNode(GraphNode{ID: sharedNode, Kind: "rule", Label: object.ID, Group: "shared", Source: firstSource(object.Sources)})
		for _, truth := range object.TruthPaths {
			addSource(truth)
			fileNode := "file:" + truth.Path
			builder.addNode(GraphNode{ID: fileNode, Kind: "truth_file", Label: compactTruthFileLabel(filepath.Base(truth.Path)), Group: "truth", Source: ptr(SourceRef{Path: truth.Path})})
			builder.addEdge(GraphEdge{ID: sharedNode + "->" + fileNode, From: sharedNode, To: fileNode, Kind: "described_by", Label: "described by", Source: ptr(truth)})
		}
		for _, bound := range object.BoundObjects {
			if strings.HasPrefix(bound, "unit:") || strings.HasPrefix(bound, "scenario:") {
				builder.addEdge(GraphEdge{ID: sharedNode + "->" + bound, From: sharedNode, To: bound, Kind: "bound_to", Label: "bound to", Source: firstSource(object.Sources)})
			}
		}
	}

	snapshot.Project.TruthFileCount = len(docs)
	snapshot.Nodes = builder.nodes()
	snapshot.Edges = builder.edges()
	snapshot.Sources = sortedSources(sourceSet)
	sortObjects(snapshot.Objects)
	normalizeSnapshotSlices(&snapshot)
	return snapshot
}

func buildObjectFromStatus(status statusfile.ObjectStatus, mapping repositoryMapping, docs map[string]markdownDoc) ObjectView {
	nextIntent := nextIntentFromStatus(status)
	object := ObjectView{
		ID:              status.Object,
		Kind:            status.ObjectType,
		Label:           status.Object,
		Layer:           status.ActiveLayer,
		HumanState:      humanLayer(status.ActiveLayer),
		Stable:          status.Stable,
		Candidate:       status.Candidate,
		NextCommand:     status.NextCommand,
		NextLabel:       humanNextCommand(status.NextCommand),
		NextIntent:      nextIntent,
		NextIntentLabel: humanNextIntent(nextIntent),
		Notes:           status.Notes,
		Sources:         []SourceRef{{Path: "docs/specs/_status.md", Label: "Status"}},
	}
	if status.ObjectType == "unit" {
		if unit, ok := mapping.Units[status.Object]; ok {
			object.Responsibility = unit.Responsibility
			object.ImplementationPaths = unit.ImplementationPaths
		}
	}
	if status.ObjectType == "unit" || status.ObjectType == "scenario" {
		if truthPath, err := specpaths.ObjectMainSpecFileRef(status.ObjectType, status.ActiveLayer, status.Object); err == nil {
			object.TruthPaths = []SourceRef{{Path: truthPath, Label: "Active Truth"}}
			object.Sources = appendSourceUnique(object.Sources, SourceRef{Path: truthPath, Label: "Active Truth"})
		}
	}
	for _, truth := range object.TruthPaths {
		doc, ok := docs[truth.Path]
		if !ok {
			continue
		}
		if object.Version == "" {
			object.Version = doc.Frontmatter.Scalars["version"]
		}
		if object.NextIntent == "" {
			object.NextIntent = nextIntentFromDoc(status, doc)
			object.NextIntentLabel = humanNextIntent(object.NextIntent)
		}
		object.RuleRefs = appendUnique(object.RuleRefs, extractRuleIDs(doc.Text)...)
		object.TruthPaths = appendSourceUnique(object.TruthPaths, appendixRefsForDoc(doc, docs)...)
	}
	sort.Strings(object.RuleRefs)
	return object
}

func appendixRefsForDoc(doc markdownDoc, docs map[string]markdownDoc) []SourceRef {
	refs := []SourceRef{}
	evidenceRef := strings.TrimSpace(doc.Frontmatter.Scalars["evidence_appendix_ref"])
	if evidenceRef != "" && evidenceRef != "none" {
		if ref, ok := resolveDocRef(doc.RelPath, evidenceRef, docs); ok && isAppendixPath(ref.Path) {
			ref.Label = "Evidence Appendix"
			refs = appendSourceUnique(refs, ref)
		}
	}
	for _, match := range markdownLinkPattern.FindAllStringSubmatch(doc.Text, -1) {
		if len(match) != 2 {
			continue
		}
		ref, ok := resolveDocRef(doc.RelPath, match[1], docs)
		if !ok || !isAppendixPath(ref.Path) {
			continue
		}
		ref.Label = "Appendix"
		refs = appendSourceUnique(refs, ref)
	}
	return refs
}

func resolveDocRef(fromPath, rawRef string, docs map[string]markdownDoc) (SourceRef, bool) {
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
	if _, ok := docs[refPath]; !ok {
		return SourceRef{}, false
	}
	return SourceRef{Path: refPath}, true
}

func isAppendixPath(path string) bool {
	return strings.Contains(path, "/appendix/")
}

func buildSharedObjects(mapping repositoryMapping, docs []markdownDoc) []ObjectView {
	objects := map[string]ObjectView{}
	for id, shared := range mapping.Rules {
		objects[id] = ObjectView{
			ID:             id,
			Kind:           "rule",
			Label:          id,
			Responsibility: shared.Responsibility,
			TruthPaths:     shared.TruthPaths,
			Sources:        append([]SourceRef{{Path: "docs/specs/repository_mapping.md", Label: "Repository Mapping"}}, shared.TruthPaths...),
		}
	}
	for _, doc := range docs {
		id := doc.Frontmatter.Scalars["rule_id"]
		if id == "" {
			continue
		}
		object := objects[id]
		object.ID = id
		object.Kind = "rule"
		object.Label = id
		object.Layer = doc.Frontmatter.Scalars["layer"]
		object.HumanState = humanLayer(object.Layer)
		object.Version = doc.Frontmatter.Scalars["rule_version"]
		object.BoundObjects = appendUnique(object.BoundObjects, doc.Frontmatter.BoundObjects...)
		object.TruthPaths = appendSourceUnique(object.TruthPaths, SourceRef{Path: doc.RelPath, Label: doc.Title})
		object.Sources = appendSourceUnique(object.Sources, SourceRef{Path: doc.RelPath, Label: doc.Title})
		objects[id] = object
	}
	result := make([]ObjectView, 0, len(objects))
	for _, object := range objects {
		sort.Strings(object.BoundObjects)
		result = append(result, object)
	}
	sortObjects(result)
	return result
}

func loadTruthDocs(repoRoot string) ([]markdownDoc, []Diagnostic) {
	root := filepath.Join(repoRoot, "docs/specs")
	diagnostics := []Diagnostic{}
	docs := []markdownDoc{}
	if _, err := os.Stat(root); err != nil {
		return docs, []Diagnostic{{Severity: "error", Message: "cannot read docs/specs: " + err.Error(), Source: &SourceRef{Path: "docs/specs"}}}
	}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		data, err := os.ReadFile(path)
		if err != nil {
			diagnostics = append(diagnostics, Diagnostic{Severity: "error", Message: "cannot read " + rel + ": " + err.Error(), Source: &SourceRef{Path: rel}})
			return nil
		}
		text := string(data)
		docs = append(docs, markdownDoc{
			RelPath:     rel,
			Title:       firstTitle(text, rel),
			Frontmatter: parseFrontmatter(text),
			Text:        text,
		})
		return nil
	})
	if err != nil {
		diagnostics = append(diagnostics, Diagnostic{Severity: "error", Message: "cannot walk docs/specs: " + err.Error(), Source: &SourceRef{Path: "docs/specs"}})
	}
	sort.Slice(docs, func(i, j int) bool { return docs[i].RelPath < docs[j].RelPath })
	return docs, diagnostics
}

func firstTitle(text, fallback string) string {
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return filepath.Base(fallback)
}

func compactTruthFileLabel(filename string) string {
	base := strings.TrimSuffix(filename, ".md")
	switch {
	case strings.HasPrefix(base, "c_unit_"):
		return strings.ReplaceAll(strings.TrimPrefix(base, "c_unit_"), "_", " ") + " (candidate)"
	case strings.HasPrefix(base, "s_unit_"):
		return strings.ReplaceAll(strings.TrimPrefix(base, "s_unit_"), "_", " ") + " (stable)"
	case strings.HasPrefix(base, "c_b_rule_"):
		return "shared " + strings.ReplaceAll(strings.TrimPrefix(base, "c_b_rule_"), "_", " ") + " (candidate)"
	case strings.HasPrefix(base, "s_b_rule_"):
		return "shared " + strings.ReplaceAll(strings.TrimPrefix(base, "s_b_rule_"), "_", " ") + " (stable)"
	default:
		return strings.ReplaceAll(base, "_", " ")
	}
}

func extractRuleIDs(text string) []string {
	matches := sharedRefPattern.FindAllStringSubmatch(text, -1)
	ids := []string{}
	for _, match := range matches {
		if len(match) != 2 {
			continue
		}
		id := sharedRefToID(match[1])
		if id != "" {
			ids = appendUnique(ids, id)
		}
	}
	return ids
}

func sharedRefToID(ref string) string {
	ref = strings.TrimSpace(ref)
	if before, _, ok := strings.Cut(ref, "@"); ok {
		ref = before
	}
	ref = strings.TrimPrefix(ref, "c_")
	ref = strings.TrimPrefix(ref, "s_")
	return ref
}

func humanLayer(layer string) string {
	switch layer {
	case "stable":
		return "已确认的设计基线"
	case "candidate":
		return "正在确认的设计"
	default:
		return ""
	}
}

func humanNextCommand(command string) string {
	switch command {
	case "unit_init":
		return "初始化能力真相"
	case "unit_new":
		return "创建新的能力设计"
	case "unit_fork":
		return "从已确认基线开启新一轮设计"
	case "unit_stable_verify":
		return "检查实现是否仍符合已确认设计"
	case "unit_check":
		return "检查设计是否足够支撑开发"
	case "unit_plan":
		return "把设计整理成开发计划"
	case "unit_impl":
		return "按计划实现"
	case "unit_verify":
		return "验证实现是否符合设计"
	case "unit_promote":
		return "把确认结果沉淀为正式基线"
	case "scenario_new":
		return "创建新的端到端流程设计"
	case "scenario_fork":
		return "从已确认流程开启新一轮设计"
	case "scenario_check":
		return "检查流程设计是否足够支撑验证"
	case "scenario_verify":
		return "验证端到端流程"
	case "scenario_promote":
		return "把流程确认结果沉淀为正式基线"
	case "scenario_stable_verify":
		return "检查端到端流程是否仍符合已确认设计"
	default:
		return command
	}
}

func nextIntentFromStatus(status statusfile.ObjectStatus) string {
	if status.ObjectType != "unit" || status.NextCommand != "unit_fork" {
		return ""
	}
	match := candidateIntentPattern.FindStringSubmatch(strings.ToLower(status.Notes))
	if len(match) != 2 {
		return ""
	}
	return match[1]
}

func nextIntentFromDoc(status statusfile.ObjectStatus, doc markdownDoc) string {
	if status.ObjectType != "unit" || status.ActiveLayer != "candidate" {
		return ""
	}
	intent := strings.ToLower(strings.TrimSpace(doc.Frontmatter.Scalars["candidate_intent"]))
	if intent == "repair" || intent == "change" {
		return intent
	}
	return ""
}

func humanNextIntent(intent string) string {
	switch intent {
	case "repair":
		return "修复基线"
	case "change":
		return "开启变更轮次"
	default:
		return ""
	}
}

func appendUnique(items []string, values ...string) []string {
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

func sortedSources(sourceSet map[string]SourceRef) []SourceRef {
	result := make([]SourceRef, 0, len(sourceSet))
	for _, source := range sourceSet {
		result = append(result, source)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}

func sortObjects(objects []ObjectView) {
	sort.Slice(objects, func(i, j int) bool {
		if objects[i].Kind != objects[j].Kind {
			return objects[i].Kind < objects[j].Kind
		}
		return objects[i].ID < objects[j].ID
	})
}

func ptr[T any](value T) *T {
	return &value
}

func firstSource(sources []SourceRef) *SourceRef {
	if len(sources) == 0 {
		return nil
	}
	return &sources[0]
}

func normalizeSnapshotSlices(snapshot *Snapshot) {
	if snapshot.Objects == nil {
		snapshot.Objects = []ObjectView{}
	}
	if snapshot.Nodes == nil {
		snapshot.Nodes = []GraphNode{}
	}
	if snapshot.Edges == nil {
		snapshot.Edges = []GraphEdge{}
	}
	if snapshot.Sources == nil {
		snapshot.Sources = []SourceRef{}
	}
	if snapshot.Diagnostics == nil {
		snapshot.Diagnostics = []Diagnostic{}
	}
	for idx := range snapshot.Objects {
		object := &snapshot.Objects[idx]
		if object.TruthPaths == nil {
			object.TruthPaths = []SourceRef{}
		}
		if object.ImplementationPaths == nil {
			object.ImplementationPaths = []SourceRef{}
		}
		if object.RuleRefs == nil {
			object.RuleRefs = []string{}
		}
		if object.BoundObjects == nil {
			object.BoundObjects = []string{}
		}
		if object.Sources == nil {
			object.Sources = []SourceRef{}
		}
	}
}

type graphBuilder struct {
	nodeMap map[string]GraphNode
	edgeMap map[string]GraphEdge
}

func newGraphBuilder() *graphBuilder {
	return &graphBuilder{nodeMap: map[string]GraphNode{}, edgeMap: map[string]GraphEdge{}}
}

func (b *graphBuilder) addNode(node GraphNode) {
	if node.ID == "" {
		return
	}
	if existing, ok := b.nodeMap[node.ID]; ok {
		if existing.Source == nil && node.Source != nil {
			existing.Source = node.Source
		}
		if existing.Label == "" {
			existing.Label = node.Label
		}
		b.nodeMap[node.ID] = existing
		return
	}
	b.nodeMap[node.ID] = node
}

func (b *graphBuilder) addEdge(edge GraphEdge) {
	if edge.ID == "" || edge.From == "" || edge.To == "" {
		return
	}
	b.edgeMap[edge.ID] = edge
}

func (b *graphBuilder) nodes() []GraphNode {
	result := make([]GraphNode, 0, len(b.nodeMap))
	for _, node := range b.nodeMap {
		result = append(result, node)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func (b *graphBuilder) edges() []GraphEdge {
	result := make([]GraphEdge, 0, len(b.edgeMap))
	for _, edge := range b.edgeMap {
		result = append(result, edge)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}
