package consul

import (
	"context"
	"strings"
	"time"

	consulapi "github.com/hashicorp/consul/api"

	"github.com/CloudRoamer/aimo-libs/config"
)

const (
	// Consul blocking query 默认等待时间
	defaultWaitTime = 5 * time.Minute
	// 错误后重试等待时间
	retryWaitTime = 5 * time.Second
)

type watcher struct {
	client    *consulapi.Client
	prefix    string
	lastIndex uint64
	stopCh    chan struct{}
	eventCh   chan config.Event
}

func newWatcher(client *consulapi.Client, prefix string) *watcher {
	return &watcher{
		client: client,
		prefix: prefix,
		stopCh: make(chan struct{}),
	}
}

func (w *watcher) Start(ctx context.Context) (<-chan config.Event, error) {
	w.eventCh = make(chan config.Event, 10)

	go w.watch(ctx)

	return w.eventCh, nil
}

func (w *watcher) watch(ctx context.Context) {
	defer close(w.eventCh)

	kv := w.client.KV()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopCh:
			return
		default:
		}

		// 使用 Consul blocking query 等待变更
		opts := &consulapi.QueryOptions{
			WaitIndex: w.lastIndex,
			WaitTime:  defaultWaitTime,
		}

		pairs, meta, err := kv.List(w.prefix, opts.WithContext(ctx))
		if err != nil {
			w.eventCh <- config.Event{
				Type:      config.EventTypeError,
				Source:    "consul",
				Timestamp: time.Now(),
				Error:     err,
			}
			// 错误后等待重试，避免频繁请求
			select {
			case <-ctx.Done():
				return
			case <-w.stopCh:
				return
			case <-time.After(retryWaitTime):
				continue
			}
		}

		// 如果索引发生变化，说明有配置更新
		if meta.LastIndex > w.lastIndex {
			w.lastIndex = meta.LastIndex

			// 收集变更的 key（转换为配置 key 格式）
			keys := make([]string, 0, len(pairs))
			prefixWithSep := w.prefix + "/"
			for _, pair := range pairs {
				key := strings.TrimPrefix(pair.Key, prefixWithSep)
				if key != "" && key != pair.Key {
					key = strings.ReplaceAll(key, "/", ".")
					keys = append(keys, key)
				}
			}

			w.eventCh <- config.Event{
				Type:      config.EventTypeReload,
				Source:    "consul",
				Keys:      keys,
				Timestamp: time.Now(),
			}
		}
	}
}

func (w *watcher) Stop() error {
	close(w.stopCh)
	return nil
}
