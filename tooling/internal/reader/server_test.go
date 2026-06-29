package reader

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func createReaderTestRepo(t *testing.T) string {
	t.Helper()
	repoRoot := t.TempDir()

	// Create the minimal framework structure for specflowlayout
	toolingDir := filepath.Join(repoRoot, "tooling")
	os.MkdirAll(toolingDir, 0755)
	os.WriteFile(filepath.Join(toolingDir, "go.mod"), []byte("module test\n"), 0644)

	// Create web assets for the reader
	webDir := filepath.Join(repoRoot, "tooling/reader/web")
	os.MkdirAll(webDir, 0755)
	for _, f := range []string{"index.html", "styles.css", "app.js", "cytoscape.min.js", "mermaid.min.js"} {
		os.WriteFile(filepath.Join(webDir, f), []byte("test"), 0644)
	}

	return repoRoot
}

func TestAPISnapshotReturnsJSON(t *testing.T) {
	repoRoot := createReaderTestRepo(t)

	candidateDir := filepath.Join(repoRoot, "docs/specs/units/candidate")
	os.MkdirAll(candidateDir, 0755)
	os.WriteFile(filepath.Join(candidateDir, "c_unit_api_test.md"), []byte("---\nid: api_test\nlayer: candidate\nversion: 1.0.0\nunit_refs: none\nrule_refs: none\n---\n# API Test\n"), 0644)

	store, err := NewStore(repoRoot)
	if err != nil {
		t.Fatal(err)
	}

	handler, err := NewHandler(store)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/snapshot")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var result Snapshot
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("json decode: %v", err)
	}

	if len(result.Objects) == 0 {
		t.Error("expected objects in snapshot")
	}
	if result.Project.RepoRoot == "" {
		t.Error("expected repo_root to be set")
	}
	if len(result.Sources) == 0 {
		t.Error("expected sources in snapshot")
	}
}

func TestAPISourceReturnsContent(t *testing.T) {
	repoRoot := createReaderTestRepo(t)

	candidateDir := filepath.Join(repoRoot, "docs/specs/units/candidate")
	os.MkdirAll(candidateDir, 0755)
	specContent := "---\nid: source_test\nlayer: candidate\nversion: 1.0.0\n---\n# Source Test\n"
	os.WriteFile(filepath.Join(candidateDir, "c_unit_source_test.md"), []byte(specContent), 0644)

	store, err := NewStore(repoRoot)
	if err != nil {
		t.Fatal(err)
	}

	handler, err := NewHandler(store)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	sourcePath := "docs/specs/units/candidate/c_unit_source_test.md"
	resp, err := http.Get(server.URL + "/api/source?path=" + sourcePath)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for source, got %d", resp.StatusCode)
	}
}

func TestAPISourceRejectsInvalidPath(t *testing.T) {
	repoRoot := createReaderTestRepo(t)

	store, err := NewStore(repoRoot)
	if err != nil {
		t.Fatal(err)
	}

	handler, err := NewHandler(store)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/source?path=../../../etc/passwd")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden && resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 403 or 400 for path traversal, got %d", resp.StatusCode)
	}
}

func TestAPISourceDiffWorks(t *testing.T) {
	repoRoot := createReaderTestRepo(t)

	candidateDir := filepath.Join(repoRoot, "docs/specs/units/candidate")
	stableDir := filepath.Join(repoRoot, "docs/specs/units/stable")
	os.MkdirAll(candidateDir, 0755)
	os.MkdirAll(stableDir, 0755)

	os.WriteFile(filepath.Join(candidateDir, "c_unit_diff_test.md"), []byte("---\nid: diff_test\nlayer: candidate\nversion: 2.0.0\n---\n# Diff Test\nNew content\n"), 0644)
	os.WriteFile(filepath.Join(stableDir, "s_unit_diff_test.md"), []byte("---\nid: diff_test\nlayer: stable\nversion: 1.0.0\n---\n# Diff Test\nOld content\n"), 0644)

	store, err := NewStore(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(store)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/source-diff?path=docs/specs/units/candidate/c_unit_diff_test.md")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var diff SourceDiff
	if err := json.NewDecoder(resp.Body).Decode(&diff); err != nil {
		t.Fatalf("json decode: %v", err)
	}
	if !diff.Available {
		t.Error("expected diff to be available")
	}
}

func TestSnapshotServesViaHTTP(t *testing.T) {
	repoRoot := createReaderTestRepo(t)

	candidateDir := filepath.Join(repoRoot, "docs/specs/units/candidate")
	os.MkdirAll(candidateDir, 0755)
	os.WriteFile(filepath.Join(candidateDir, "c_unit_http_test.md"), []byte("---\nid: http_test\nlayer: candidate\nversion: 1.0.0\nunit_refs: none\nrule_refs: none\n---\n# HTTP Test\n"), 0644)

	store, err := NewStore(repoRoot)
	if err != nil {
		t.Fatal(err)
	}
	handler, err := NewHandler(store)
	if err != nil {
		t.Fatal(err)
	}

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/snapshot")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var snap Snapshot
	json.NewDecoder(resp.Body).Decode(&snap)

	if len(snap.Objects) == 0 {
		t.Fatal("no objects in snapshot")
	}

	var obj *ObjectView
	for _, o := range snap.Objects {
		if o.ID == "http_test" {
			obj = &o
			break
		}
	}
	if obj == nil {
		t.Fatal("http_test not found")
	}
	if !obj.HasCandidate {
		t.Error("expected candidate file to be detected")
	}
}

func TestStoreRefreshUpdatesVersion(t *testing.T) {
	repoRoot := createReaderTestRepo(t)

	store, err := NewStore(repoRoot)
	if err != nil {
		t.Fatal(err)
	}

	snap1 := store.RefreshSnapshot()
	snap2 := store.RefreshSnapshot()

	if snap1.Version > snap2.Version {
		t.Error("expected non-decreasing version")
	}
}
