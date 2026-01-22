package file

import (
	"context"
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

	if err := fsWatcher.Add(w.path); err != nil {
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

			// 只关注写入和创建事件
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
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
