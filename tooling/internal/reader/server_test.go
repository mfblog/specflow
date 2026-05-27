package reader

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSnapshotAndSourceEndpoints(t *testing.T) {
	repoRoot := createReaderRepo(t)
	createReaderWeb(t, repoRoot)
	store, err := NewStore(repoRoot)
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	handler, err := NewHandler(store)
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/snapshot", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var snapshot Snapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("snapshot json invalid: %v", err)
	}
	if snapshot.Version != 2 {
		t.Fatalf("expected snapshot endpoint to refresh to version 2, got %d", snapshot.Version)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/source?path=docs/specs/_status.md", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/source?path=AGENTS.md", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for disallowed source, got %d", rec.Code)
	}

	writeReaderTestFile(t, filepath.Join(repoRoot, "docs/specs/units/stable/s_unit_memory.md"), "# Memory\n\nStable line.\n")
	req = httptest.NewRequest(http.MethodGet, "/api/source-diff?path=docs/specs/units/candidate/c_unit_memory.md", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected source diff 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var diff SourceDiff
	if err := json.Unmarshal(rec.Body.Bytes(), &diff); err != nil {
		t.Fatalf("source diff json invalid: %v", err)
	}
	if !diff.Available || diff.StablePath != "docs/specs/units/stable/s_unit_memory.md" {
		t.Fatalf("unexpected source diff: %+v", diff)
	}

	req = httptest.NewRequest(http.MethodGet, "/api/source-diff?path=../AGENTS.md", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for escaped source diff path, got %d", rec.Code)
	}
}

func TestSnapshotEndpointRefreshesFromDisk(t *testing.T) {
	repoRoot := createReaderRepo(t)
	createReaderWeb(t, repoRoot)
	store, err := NewStore(repoRoot)
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	handler, err := NewHandler(store)
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	statusPath := filepath.Join(repoRoot, "docs/specs/_status.md")
	data, err := os.ReadFile(statusPath)
	if err != nil {
		t.Fatalf("ReadFile(status) failed: %v", err)
	}
	updated := strings.Replace(string(data), "`unit_check` | note", "`unit_impl` | note", 1)
	if updated == string(data) {
		t.Fatalf("test fixture did not contain the expected status row")
	}
	if err := os.WriteFile(statusPath, []byte(updated), 0o644); err != nil {
		t.Fatalf("WriteFile(status) failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/snapshot", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var snapshot Snapshot
	if err := json.Unmarshal(rec.Body.Bytes(), &snapshot); err != nil {
		t.Fatalf("snapshot json invalid: %v", err)
	}
	unit := findObject(t, snapshot.Objects, "unit", "assistant")
	if unit.NextCommand != "unit_impl" {
		t.Fatalf("expected refreshed next command unit_impl, got %q", unit.NextCommand)
	}
}

func TestStaticWebFilesAreServedFromDisk(t *testing.T) {
	repoRoot := createReaderRepo(t)
	createReaderWeb(t, repoRoot)
	store, err := NewStore(repoRoot)
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	handler, err := NewHandler(store)
	if err != nil {
		t.Fatalf("NewHandler returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/app.js", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "disk asset") {
		t.Fatalf("expected app.js to be served from disk, got %q", rec.Body.String())
	}
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("expected static files to disable browser cache, got %q", rec.Header().Get("Cache-Control"))
	}
}

func TestNewHandlerRequiresReaderWebAssets(t *testing.T) {
	repoRoot := createReaderRepo(t)
	store, err := NewStore(repoRoot)
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	_, err = NewHandler(store)
	if err == nil || !strings.Contains(err.Error(), "reader web root missing") {
		t.Fatalf("expected missing web root error, got %v", err)
	}

	createReaderWeb(t, repoRoot)
	if err := os.Remove(filepath.Join(repoRoot, "specflow/tooling/reader/web/app.js")); err != nil {
		t.Fatalf("Remove(app.js) failed: %v", err)
	}
	_, err = NewHandler(store)
	if err == nil || !strings.Contains(err.Error(), "reader web asset missing: specflow/tooling/reader/web/app.js") {
		t.Fatalf("expected missing app.js error, got %v", err)
	}
}

func TestReaderWebRootSupportsSourceRepoLayout(t *testing.T) {
	repoRoot := t.TempDir()
	writeReaderTestFile(t, filepath.Join(repoRoot, "tooling/manifest.tsv"), "templates/AGENTS.md\tAGENTS.md\tframework\n")
	createReaderWebAt(t, repoRoot, "tooling/reader/web")

	webRoot, err := ReaderWebRoot(repoRoot)
	if err != nil {
		t.Fatalf("ReaderWebRoot returned error: %v", err)
	}
	want := filepath.Join(repoRoot, "tooling/reader/web")
	if webRoot != want {
		t.Fatalf("ReaderWebRoot = %q, want %q", webRoot, want)
	}
}

func createReaderWeb(t *testing.T, repoRoot string) {
	t.Helper()
	createReaderWebAt(t, repoRoot, "specflow/tooling/reader/web")
}

func createReaderWebAt(t *testing.T, repoRoot, relativeRoot string) {
	t.Helper()
	webRoot := filepath.Join(repoRoot, filepath.FromSlash(relativeRoot))
	writeReaderTestFile(t, filepath.Join(webRoot, "index.html"), "<!doctype html><script src=\"/app.js\"></script>\n")
	writeReaderTestFile(t, filepath.Join(webRoot, "styles.css"), "body { color: #111; }\n")
	writeReaderTestFile(t, filepath.Join(webRoot, "app.js"), "console.log('disk asset');\n")
	writeReaderTestFile(t, filepath.Join(webRoot, "cytoscape.min.js"), "window.cytoscape = function() {};\n")
	writeReaderTestFile(t, filepath.Join(webRoot, "mermaid.min.js"), "window.mermaid = { initialize() {}, run() {} };\n")
}
