package reader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var requiredWebFiles = []string{
	"index.html",
	"styles.css",
	"app.js",
	"cytoscape.min.js",
	"mermaid.min.js",
}

type ServeOptions struct {
	RepoRoot string
	Addr     string
}

func Serve(ctx context.Context, options ServeOptions, stdout io.Writer) error {
	if strings.TrimSpace(options.Addr) == "" {
		options.Addr = "127.0.0.1:17863"
	}
	store, err := NewStore(options.RepoRoot)
	if err != nil {
		return err
	}
	handler, err := NewHandler(store)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              options.Addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}
	listener, err := net.Listen("tcp", options.Addr)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "specFlow reader serving at http://%s\n", listener.Addr().String())
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()
	err = server.Serve(listener)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func NewHandler(store *Store) (http.Handler, error) {
	webRoot, err := ReaderWebRoot(store.RepoRoot())
	if err != nil {
		return nil, err
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/snapshot", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, store.RefreshSnapshot())
	})
	mux.HandleFunc("/api/source", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		source, err := ReadAllowedSource(store.RepoRoot(), r.URL.Query().Get("path"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, source)
	})
	mux.HandleFunc("/api/source-diff", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		diff, err := ReadAllowedSourceDiff(store.RepoRoot(), r.URL.Query().Get("path"))
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, diff)
	})
	mux.Handle("/", noStore(http.FileServer(http.Dir(webRoot))))
	return mux, nil
}

func noStore(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

func ReaderWebRoot(repoRoot string) (string, error) {
	webRoot := filepath.Join(repoRoot, "specflow", "tooling", "reader", "web")
	info, err := os.Stat(webRoot)
	if err != nil {
		return "", fmt.Errorf("reader web root missing: %s", filepath.ToSlash(filepath.Join("specflow", "tooling", "reader", "web")))
	}
	if !info.IsDir() {
		return "", fmt.Errorf("reader web root is not a directory: %s", filepath.ToSlash(filepath.Join("specflow", "tooling", "reader", "web")))
	}
	for _, file := range requiredWebFiles {
		path := filepath.Join(webRoot, file)
		info, err := os.Stat(path)
		if err != nil {
			return "", fmt.Errorf("reader web asset missing: %s", filepath.ToSlash(filepath.Join("specflow", "tooling", "reader", "web", file)))
		}
		if info.IsDir() {
			return "", fmt.Errorf("reader web asset is a directory: %s", filepath.ToSlash(filepath.Join("specflow", "tooling", "reader", "web", file)))
		}
	}
	return webRoot, nil
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	_ = encoder.Encode(value)
}
