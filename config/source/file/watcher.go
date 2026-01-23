package file

import (
	"context"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"

	"github.com/CloudRoamer/aimo-libs/config"
)

// watcher 文件监听器
type watcher struct {
	path    string
	stopCh  chan struct{}
	eventCh chan config.Event
}

func newWatcher(path string) *watcher {
	if absPath, err := filepath.Abs(path); err == nil {
		path = absPath
	}

	return &watcher{
		path:   path,
		stopCh: make(chan struct{}),
	}
}

func (w *watcher) Start(ctx context.Context) (<-chan config.Event, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	watchDir := filepath.Dir(w.path)
	if err := fsWatcher.Add(watchDir); err != nil {
		fsWatcher.Close()
		return nil, err
	}

	w.eventCh = make(chan config.Event, 10)

	go w.watch(ctx, fsWatcher)

	return w.eventCh, nil
}

func (w *watcher) watch(ctx context.Context, fsWatcher *fsnotify.Watcher) {
	defer close(w.eventCh)
	defer fsWatcher.Close()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		case event, ok := <-fsWatcher.Events:
			if !ok {
				return
			}

			if filepath.Clean(event.Name) != w.path {
				continue
			}

			if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename|fsnotify.Remove|fsnotify.Chmod) != 0 {
				w.eventCh <- config.Event{
					Type:      config.EventTypeUpdate,
					Source:    "file:" + w.path,
					Timestamp: time.Now(),
				}
			}
		case err, ok := <-fsWatcher.Errors:
			if !ok {
				return
			}
			w.eventCh <- config.Event{
				Type:      config.EventTypeError,
				Source:    "file:" + w.path,
				Timestamp: time.Now(),
				Error:     err,
			}
		}
	}
}

func (w *watcher) Stop() error {
	close(w.stopCh)
	return nil
}
