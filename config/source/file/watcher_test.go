package file

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CloudRoamer/aimo-libs/config"
)

func TestWatcher_StartAndStop(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "config.json")

	if err := os.WriteFile(path, []byte("{}"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	watcher := newWatcher(path)
	ch, err := watcher.Start(ctx)
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if err := os.WriteFile(path, []byte("{\"updated\":true}"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	select {
	case event := <-ch:
		if event.Type != config.EventTypeUpdate {
			t.Fatalf("event.Type = %v, want %v", event.Type, config.EventTypeUpdate)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for update event")
	}

	if err := watcher.Stop(); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}
	cancel()

	closed := make(chan struct{})
	go func() {
		for range ch {
		}
		close(closed)
	}()

	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("event channel not closed after stop")
	}
}
