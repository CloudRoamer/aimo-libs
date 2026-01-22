package config

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// Manager 配置管理器
// 负责协调多个配置源、执行合并、处理热更新
type Manager struct {
	mu       sync.RWMutex
	sources  []Source
	merger   Merger
	config   *configImpl
	watchers []Watcher
	onChange []ChangeCallback
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// ChangeCallback 配置变更回调函数
// oldConfig 可能为 nil（首次加载时）
type ChangeCallback func(event Event, oldConfig, newConfig Config)

// NewManager 创建配置管理器
func NewManager(opts ...ManagerOption) *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	m := &Manager{
		sources:  make([]Source, 0),
		merger:   NewDefaultMerger(),
		config:   newConfigImpl(),
		onChange: make([]ChangeCallback, 0),
		ctx:      ctx,
		cancel:   cancel,
	}

	for _, opt := range opts {
		opt(m)
	}

	return m
}

// AddSource 添加配置源
// 配置源按优先级排序，高优先级的源在合并时覆盖低优先级
func (m *Manager) AddSource(sources ...Source) *Manager {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sources = append(m.sources, sources...)
	// 按优先级排序（从低到高，便于后续合并时高优先级覆盖低优先级）
	sort.Slice(m.sources, func(i, j int) bool {
		return m.sources[i].Priority() < m.sources[j].Priority()
	})

	return m
}

// Load 加载所有配置源并合并
func (m *Manager) Load(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.loadLocked(ctx)
}

// loadLocked 内部加载方法（需要持有锁）
func (m *Manager) loadLocked(ctx context.Context) error {
	allValues := make([]map[string]Value, 0, len(m.sources))

	for _, source := range m.sources {
		values, err := source.Load(ctx)
		if err != nil {
			return fmt.Errorf("failed to load from source %s: %w", source.Name(), err)
		}
		allValues = append(allValues, values)
	}

	// 合并所有配置（按优先级，后面的覆盖前面的）
	merged := m.merger.Merge(allValues...)
	m.config = newConfigImplFromMap(merged)

	return nil
}

// Watch 启动所有配置源的监听
func (m *Manager) Watch() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, source := range m.sources {
		watcher := source.Watch()
		if watcher == nil {
			continue
		}

		eventCh, err := watcher.Start(m.ctx)
		if err != nil {
			return fmt.Errorf("failed to start watcher for source %s: %w", source.Name(), err)
		}

		m.watchers = append(m.watchers, watcher)

		// 启动 goroutine 处理事件
		m.wg.Add(1)
		go func(sourceName string, eventCh <-chan Event) {
			defer m.wg.Done()
			m.handleEvents(sourceName, eventCh)
		}(source.Name(), eventCh)
	}

	return nil
}

// handleEvents 处理配置变更事件
func (m *Manager) handleEvents(sourceName string, eventCh <-chan Event) {
	for {
		select {
		case <-m.ctx.Done():
			return
		case event, ok := <-eventCh:
			if !ok {
				return
			}

			if event.Type == EventTypeError {
				// 记录错误但不停止监听
				m.notifyChange(event, nil, nil)
				continue
			}

			// 重新加载配置
			m.mu.Lock()
			oldConfig := m.config.clone()
			if err := m.loadLocked(m.ctx); err != nil {
				m.mu.Unlock()
				m.notifyChange(Event{
					Type:      EventTypeError,
					Source:    sourceName,
					Timestamp: event.Timestamp,
					Error:     err,
				}, nil, nil)
				continue
			}
			newConfig := m.config.clone()
			m.mu.Unlock()

			m.notifyChange(event, oldConfig, newConfig)
		}
	}
}

// OnChange 注册配置变更回调
func (m *Manager) OnChange(callback ChangeCallback) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onChange = append(m.onChange, callback)
}

// notifyChange 通知所有回调
func (m *Manager) notifyChange(event Event, oldConfig, newConfig Config) {
	m.mu.RLock()
	callbacks := make([]ChangeCallback, len(m.onChange))
	copy(callbacks, m.onChange)
	m.mu.RUnlock()

	for _, cb := range callbacks {
		cb(event, oldConfig, newConfig)
	}
}

// Config 获取当前配置（只读）
func (m *Manager) Config() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.config
}

// Close 关闭管理器，停止所有监听
func (m *Manager) Close() error {
	m.cancel()

	// 停止所有 watchers
	var errs []error
	for _, w := range m.watchers {
		if err := w.Stop(); err != nil {
			errs = append(errs, err)
		}
	}

	// 等待所有 goroutine 退出
	m.wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("errors while closing watchers: %v", errs)
	}
	return nil
}

// clone 克隆配置（用于热更新时比较）
func (c *configImpl) clone() *configImpl {
	newData := make(map[string]Value, len(c.data))
	for k, v := range c.data {
		newData[k] = v
	}
	return &configImpl{data: newData}
}

// NewDefaultMerger 创建默认合并器
func NewDefaultMerger() Merger {
	return &defaultMerger{}
}

// defaultMerger 默认合并器实现
type defaultMerger struct{}

func (m *defaultMerger) Merge(maps ...map[string]Value) map[string]Value {
	result := make(map[string]Value)

	for _, configMap := range maps {
		for k, v := range configMap {
			result[k] = v
		}
	}

	return result
}
