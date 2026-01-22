package file

import (
	"context"

	"github.com/CloudRoamer/aimo-libs/config"
)

// watcher 文件监听器（本阶段暂不实现）
type watcher struct {
	path string
}

func newWatcher(path string) *watcher {
	return &watcher{
		path: path,
	}
}

func (w *watcher) Start(ctx context.Context) (<-chan config.Event, error) {
	// 本阶段暂不实现
	return nil, nil
}

func (w *watcher) Stop() error {
	// 本阶段暂不实现
	return nil
}
