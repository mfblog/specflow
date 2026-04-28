package reader

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
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
	if snapshot.Version != 1 {
		t.Fatalf("expected initial snapshot version 1, got %d", snapshot.Version)
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

func TestEventsEndpointEmitsInitialVersion(t *testing.T) {
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

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/events", nil).WithContext(ctx)
	rec := &safeRecorder{ResponseRecorder: httptest.NewRecorder()}
	done := make(chan struct{})
	go func() {
		handler.ServeHTTP(rec, req)
		close(done)
	}()

	deadline := time.After(time.Second)
	for {
		if strings.Contains(rec.BodyString(), "event: snapshot") {
			cancel()
			<-done
			return
		}
		select {
		case <-deadline:
			cancel()
			<-done
			t.Fatalf("expected snapshot event, got %q", rec.BodyString())
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}

type safeRecorder struct {
	*httptest.ResponseRecorder
	mu sync.Mutex
}

func (r *safeRecorder) Write(data []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.ResponseRecorder.Write(data)
}

func (r *safeRecorder) Flush() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.ResponseRecorder.Flush()
}

func (r *safeRecorder) BodyString() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Body.String()
}

func createReaderWeb(t *testing.T, repoRoot string) {
	t.Helper()
	webRoot := filepath.Join(repoRoot, "specflow/tooling/reader/web")
	writeReaderTestFile(t, filepath.Join(webRoot, "index.html"), "<!doctype html><script src=\"/app.js\"></script>\n")
	writeReaderTestFile(t, filepath.Join(webRoot, "styles.css"), "body { color: #111; }\n")
	writeReaderTestFile(t, filepath.Join(webRoot, "app.js"), "console.log('disk asset');\n")
	writeReaderTestFile(t, filepath.Join(webRoot, "cytoscape.min.js"), "window.cytoscape = function() {};\n")
	writeReaderTestFile(t, filepath.Join(webRoot, "mermaid.min.js"), "window.mermaid = { initialize() {}, run() {} };\n")
}
