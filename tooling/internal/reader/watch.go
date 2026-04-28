package reader

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

func WatchSpecs(ctx context.Context, store *Store) error {
	specsRoot := filepath.Join(store.RepoRoot(), "docs", "specs")
	if _, err := os.Stat(specsRoot); err != nil {
		return nil
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	if err := addWatchTree(watcher, specsRoot); err != nil {
		_ = watcher.Close()
		return err
	}

	go func() {
		defer watcher.Close()
		timer := time.NewTimer(time.Hour)
		if !timer.Stop() {
			<-timer.C
		}
		pending := false
		for {
			select {
			case <-ctx.Done():
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Create == fsnotify.Create {
					if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
						_ = addWatchTree(watcher, event.Name)
					}
				}
				if !isSpecFileEvent(event.Name) {
					continue
				}
				if !pending {
					timer.Reset(150 * time.Millisecond)
					pending = true
				}
			case <-watcher.Errors:
			case <-timer.C:
				pending = false
				store.Rebuild()
			}
		}
	}()
	return nil
}

func addWatchTree(watcher *fsnotify.Watcher, root string) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		return watcher.Add(path)
	})
}

func isSpecFileEvent(path string) bool {
	path = strings.ToLower(filepath.ToSlash(path))
	if strings.HasSuffix(path, ".md") {
		return true
	}
	return strings.Contains(path, "/docs/specs/")
}
