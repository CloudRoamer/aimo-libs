package config

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// mockSource 测试用的 mock 配置源
type mockSource struct {
	name     string
	priority int
	data     map[string]Value
	loadErr  error
	watcher  Watcher
}

func (m *mockSource) Name() string {
	return m.name
}

func (m *mockSource) Priority() int {
	return m.priority
}

func (m *mockSource) Load(ctx context.Context) (map[string]Value, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}
	return m.data, nil
}

func (m *mockSource) Watch() Watcher {
	return m.watcher
}

// mockWatcher 测试用的 mock watcher
type mockWatcher struct {
	eventCh chan Event
	stopErr error
}

func (m *mockWatcher) Start(ctx context.Context) (<-chan Event, error) {
	if m.eventCh == nil {
		m.eventCh = make(chan Event, 10)
	}
	return m.eventCh, nil
}

func (m *mockWatcher) Stop() error {
	if m.eventCh != nil {
		close(m.eventCh)
	}
	return m.stopErr
}

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	if m.sources == nil {
		t.Error("Manager sources should be initialized")
	}
	if m.merger == nil {
		t.Error("Manager merger should be initialized")
	}
	if m.config == nil {
		t.Error("Manager config should be initialized")
	}
}

func TestManager_AddSource(t *testing.T) {
	m := NewManager()

	source1 := &mockSource{name: "source1", priority: 50}
	source2 := &mockSource{name: "source2", priority: 100}
	source3 := &mockSource{name: "source3", priority: 75}

	m.AddSource(source1, source2, source3)

	if len(m.sources) != 3 {
		t.Errorf("Expected 3 sources, got %d", len(m.sources))
	}

	// 验证按优先级排序（从低到高）
	if m.sources[0].Priority() != 50 {
		t.Error("Sources should be sorted by priority")
	}
	if m.sources[1].Priority() != 75 {
		t.Error("Sources should be sorted by priority")
	}
	if m.sources[2].Priority() != 100 {
		t.Error("Sources should be sorted by priority")
	}
}

func TestManager_Load(t *testing.T) {
	t.Run("single_source", func(t *testing.T) {
		m := NewManager()
		source := &mockSource{
			name:     "test",
			priority: 50,
			data: map[string]Value{
				"key1": NewValue("value1"),
				"key2": NewValueFromInterface(42),
			},
		}
		m.AddSource(source)

		ctx := context.Background()
		err := m.Load(ctx)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		cfg := m.Config()
		if cfg.GetString("key1", "") != "value1" {
			t.Error("Config should contain loaded data")
		}
		if cfg.GetInt("key2", 0) != 42 {
			t.Error("Config should contain loaded data")
		}
	})

	t.Run("multiple_sources_priority", func(t *testing.T) {
		m := NewManager()

		lowPriority := &mockSource{
			name:     "low",
			priority: 50,
			data: map[string]Value{
				"key":    NewValue("low_value"),
				"unique": NewValue("low_unique"),
			},
		}

		highPriority := &mockSource{
			name:     "high",
			priority: 100,
			data: map[string]Value{
				"key": NewValue("high_value"), // 应覆盖 low_value
			},
		}

		m.AddSource(lowPriority, highPriority)

		ctx := context.Background()
		err := m.Load(ctx)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		cfg := m.Config()
		if cfg.GetString("key", "") != "high_value" {
			t.Error("High priority source should override low priority")
		}
		if cfg.GetString("unique", "") != "low_unique" {
			t.Error("Unique key from low priority should be preserved")
		}
	})

	t.Run("load_error", func(t *testing.T) {
		m := NewManager()
		source := &mockSource{
			name:     "failing",
			priority: 50,
			loadErr:  errors.New("load failed"),
		}
		m.AddSource(source)

		ctx := context.Background()
		err := m.Load(ctx)
		if err == nil {
			t.Error("Load() should return error when source fails")
		}
	})

	t.Run("empty_sources", func(t *testing.T) {
		m := NewManager()
		ctx := context.Background()
		err := m.Load(ctx)
		if err != nil {
			t.Errorf("Load() with no sources should not error, got %v", err)
		}
	})
}

func TestManager_Watch(t *testing.T) {
	t.Run("start_watchers", func(t *testing.T) {
		m := NewManager()

		watcher := &mockWatcher{}
		source := &mockSource{
			name:     "test",
			priority: 50,
			data:     map[string]Value{"key": NewValue("value")},
			watcher:  watcher,
		}
		m.AddSource(source)

		// 先加载配置
		ctx := context.Background()
		if err := m.Load(ctx); err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// 启动监听
		err := m.Watch()
		if err != nil {
			t.Fatalf("Watch() error = %v", err)
		}

		if len(m.watchers) != 1 {
			t.Errorf("Expected 1 watcher, got %d", len(m.watchers))
		}

		// 清理
		m.Close()
	})

	t.Run("handle_events", func(t *testing.T) {
		m := NewManager()

		watcher := &mockWatcher{
			eventCh: make(chan Event, 10),
		}
		source := &mockSource{
			name:     "test",
			priority: 50,
			data:     map[string]Value{"key": NewValue("old_value")},
			watcher:  watcher,
		}
		m.AddSource(source)

		// 先加载配置
		ctx := context.Background()
		if err := m.Load(ctx); err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// 注册回调
		var mu sync.Mutex
		var receivedEvent Event
		var callbackCalled bool

		m.OnChange(func(event Event, oldConfig, newConfig Config) {
			mu.Lock()
			defer mu.Unlock()
			receivedEvent = event
			callbackCalled = true
		})

		// 启动监听
		if err := m.Watch(); err != nil {
			t.Fatalf("Watch() error = %v", err)
		}

		// 发送事件
		watcher.eventCh <- Event{
			Type:      EventTypeUpdate,
			Source:    "test",
			Timestamp: time.Now(),
		}

		// 等待回调被调用
		time.Sleep(100 * time.Millisecond)

		mu.Lock()
		if !callbackCalled {
			t.Error("Callback should be called")
		}
		if receivedEvent.Type != EventTypeUpdate {
			t.Errorf("Expected EventTypeUpdate, got %v", receivedEvent.Type)
		}
		mu.Unlock()

		// 清理
		m.Close()
	})

	t.Run("no_watcher", func(t *testing.T) {
		m := NewManager()

		source := &mockSource{
			name:     "test",
			priority: 50,
			data:     map[string]Value{"key": NewValue("value")},
			watcher:  nil, // 不支持 watch
		}
		m.AddSource(source)

		ctx := context.Background()
		if err := m.Load(ctx); err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		err := m.Watch()
		if err != nil {
			t.Errorf("Watch() should not error when source has no watcher")
		}

		if len(m.watchers) != 0 {
			t.Error("Should not add watcher when source returns nil")
		}

		m.Close()
	})
}

func TestManager_OnChange(t *testing.T) {
	m := NewManager()

	var mu sync.Mutex
	callCount := 0

	callback := func(event Event, oldConfig, newConfig Config) {
		mu.Lock()
		defer mu.Unlock()
		callCount++
	}

	m.OnChange(callback)
	m.OnChange(callback) // 注册两次

	// 模拟事件
	m.notifyChange(Event{Type: EventTypeUpdate}, nil, nil)

	mu.Lock()
	if callCount != 2 {
		t.Errorf("Expected 2 callbacks, got %d", callCount)
	}
	mu.Unlock()
}

func TestManager_Config(t *testing.T) {
	m := NewManager()
	source := &mockSource{
		name:     "test",
		priority: 50,
		data:     map[string]Value{"key": NewValue("value")},
	}
	m.AddSource(source)

	ctx := context.Background()
	if err := m.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	cfg := m.Config()
	if cfg == nil {
		t.Fatal("Config() returned nil")
	}

	if cfg.GetString("key", "") != "value" {
		t.Error("Config() should return current config")
	}
}

func TestManager_Close(t *testing.T) {
	t.Run("close_without_watchers", func(t *testing.T) {
		m := NewManager()
		err := m.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	t.Run("close_with_watchers", func(t *testing.T) {
		m := NewManager()

		watcher := &mockWatcher{
			eventCh: make(chan Event, 10),
		}
		source := &mockSource{
			name:     "test",
			priority: 50,
			data:     map[string]Value{"key": NewValue("value")},
			watcher:  watcher,
		}
		m.AddSource(source)

		ctx := context.Background()
		if err := m.Load(ctx); err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if err := m.Watch(); err != nil {
			t.Fatalf("Watch() error = %v", err)
		}

		err := m.Close()
		if err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})

	t.Run("close_with_error", func(t *testing.T) {
		m := NewManager()

		watcher := &mockWatcher{
			eventCh: make(chan Event, 10),
			stopErr: errors.New("stop error"),
		}
		source := &mockSource{
			name:     "test",
			priority: 50,
			data:     map[string]Value{"key": NewValue("value")},
			watcher:  watcher,
		}
		m.AddSource(source)

		ctx := context.Background()
		if err := m.Load(ctx); err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		if err := m.Watch(); err != nil {
			t.Fatalf("Watch() error = %v", err)
		}

		err := m.Close()
		if err == nil {
			t.Error("Close() should return error when watcher fails")
		}
	})
}

func TestConfigImpl_Clone(t *testing.T) {
	original := newConfigImplFromMap(map[string]Value{
		"key1": NewValue("value1"),
		"key2": NewValueFromInterface(42),
	})

	cloned := original.clone()

	// 验证克隆的内容
	if cloned.GetString("key1", "") != "value1" {
		t.Error("Clone should have same data")
	}
	if cloned.GetInt("key2", 0) != 42 {
		t.Error("Clone should have same data")
	}

	// 验证是独立副本
	if &original.data == &cloned.data {
		t.Error("Clone should create a new map")
	}
}

func TestManager_ConcurrentAccess(t *testing.T) {
	m := NewManager()
	source := &mockSource{
		name:     "test",
		priority: 50,
		data:     map[string]Value{"key": NewValue("value")},
	}
	m.AddSource(source)

	ctx := context.Background()
	if err := m.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				cfg := m.Config()
				_ = cfg.GetString("key", "")
			}
		}()
	}

	wg.Wait()
}
