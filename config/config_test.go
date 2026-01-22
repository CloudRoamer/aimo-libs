package config

import (
	"sync"
	"testing"
	"time"
)

func TestConfigImpl_Get(t *testing.T) {
	cfg := newConfigImplFromMap(map[string]Value{
		"key1": NewValue("value1"),
		"key2": NewValueFromInterface(42),
	})

	t.Run("existing_key", func(t *testing.T) {
		val, ok := cfg.Get("key1")
		if !ok {
			t.Error("Get() should return true for existing key")
		}
		if val.String() != "value1" {
			t.Errorf("Get() = %v, want value1", val.String())
		}
	})

	t.Run("non_existing_key", func(t *testing.T) {
		_, ok := cfg.Get("nonexistent")
		if ok {
			t.Error("Get() should return false for non-existing key")
		}
	})
}

func TestConfigImpl_GetString(t *testing.T) {
	cfg := newConfigImplFromMap(map[string]Value{
		"existing": NewValue("value"),
	})

	tests := []struct {
		name       string
		key        string
		defaultVal string
		want       string
	}{
		{"existing_key", "existing", "default", "value"},
		{"non_existing_key", "missing", "fallback", "fallback"},
		{"empty_default", "missing", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.GetString(tt.key, tt.defaultVal); got != tt.want {
				t.Errorf("GetString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigImpl_GetInt(t *testing.T) {
	cfg := newConfigImplFromMap(map[string]Value{
		"valid":   NewValueFromInterface(42),
		"invalid": NewValue("not_a_number"),
	})

	tests := []struct {
		name       string
		key        string
		defaultVal int
		want       int
	}{
		{"valid_int", "valid", 0, 42},
		{"invalid_int", "invalid", 99, 99},
		{"missing_key", "missing", 100, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.GetInt(tt.key, tt.defaultVal); got != tt.want {
				t.Errorf("GetInt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigImpl_GetBool(t *testing.T) {
	cfg := newConfigImplFromMap(map[string]Value{
		"true_bool":   NewValueFromInterface(true),
		"false_bool":  NewValueFromInterface(false),
		"true_string": NewValue("true"),
		"invalid":     NewValue("not_bool"),
	})

	tests := []struct {
		name       string
		key        string
		defaultVal bool
		want       bool
	}{
		{"true_value", "true_bool", false, true},
		{"false_value", "false_bool", true, false},
		{"true_string", "true_string", false, true},
		{"invalid", "invalid", true, true},
		{"missing", "missing", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.GetBool(tt.key, tt.defaultVal); got != tt.want {
				t.Errorf("GetBool() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigImpl_GetDuration(t *testing.T) {
	cfg := newConfigImplFromMap(map[string]Value{
		"duration": NewValueFromInterface(30 * time.Second),
		"string":   NewValue("1m30s"),
		"invalid":  NewValue("not_duration"),
	})

	tests := []struct {
		name       string
		key        string
		defaultVal time.Duration
		want       time.Duration
	}{
		{"duration_type", "duration", 0, 30 * time.Second},
		{"string_duration", "string", 0, 90 * time.Second},
		{"invalid_duration", "invalid", 5 * time.Minute, 5 * time.Minute},
		{"missing_key", "missing", time.Hour, time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cfg.GetDuration(tt.key, tt.defaultVal); got != tt.want {
				t.Errorf("GetDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfigImpl_GetStringSlice(t *testing.T) {
	cfg := newConfigImplFromMap(map[string]Value{
		"slice": NewValueFromInterface([]string{"a", "b", "c"}),
		"json":  NewValue(`["x","y"]`),
		"comma": NewValue("foo,bar,baz"),
	})

	tests := []struct {
		name       string
		key        string
		defaultVal []string
		want       []string
	}{
		{"slice_type", "slice", nil, []string{"a", "b", "c"}},
		{"json_array", "json", nil, []string{"x", "y"}},
		{"comma_separated", "comma", nil, []string{"foo", "bar", "baz"}},
		{"missing_key", "missing", []string{"default"}, []string{"default"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetStringSlice(tt.key, tt.defaultVal)
			if len(got) != len(tt.want) {
				t.Errorf("GetStringSlice() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for i, v := range got {
				if v != tt.want[i] {
					t.Errorf("GetStringSlice()[%d] = %v, want %v", i, v, tt.want[i])
				}
			}
		})
	}
}

func TestConfigImpl_GetStringMap(t *testing.T) {
	cfg := newConfigImplFromMap(map[string]Value{
		"map":  NewValueFromInterface(map[string]string{"key": "value"}),
		"json": NewValue(`{"foo":"bar"}`),
	})

	tests := []struct {
		name       string
		key        string
		defaultVal map[string]string
		want       map[string]string
	}{
		{"map_type", "map", nil, map[string]string{"key": "value"}},
		{"json_object", "json", nil, map[string]string{"foo": "bar"}},
		{"missing_key", "missing", map[string]string{"default": "val"}, map[string]string{"default": "val"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetStringMap(tt.key, tt.defaultVal)
			if len(got) != len(tt.want) {
				t.Errorf("GetStringMap() length = %v, want %v", len(got), len(tt.want))
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("GetStringMap()[%s] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestConfigImpl_Keys(t *testing.T) {
	cfg := newConfigImplFromMap(map[string]Value{
		"key1": NewValue("val1"),
		"key2": NewValue("val2"),
		"key3": NewValue("val3"),
	})

	keys := cfg.Keys()
	if len(keys) != 3 {
		t.Errorf("Keys() returned %d keys, want 3", len(keys))
	}

	keyMap := make(map[string]bool)
	for _, k := range keys {
		keyMap[k] = true
	}

	expected := []string{"key1", "key2", "key3"}
	for _, k := range expected {
		if !keyMap[k] {
			t.Errorf("Keys() missing key: %s", k)
		}
	}
}

func TestConfigImpl_Has(t *testing.T) {
	cfg := newConfigImplFromMap(map[string]Value{
		"existing": NewValue("value"),
	})

	if !cfg.Has("existing") {
		t.Error("Has() should return true for existing key")
	}

	if cfg.Has("nonexistent") {
		t.Error("Has() should return false for non-existing key")
	}
}

func TestConfigImpl_EmptyConfig(t *testing.T) {
	cfg := newConfigImpl()

	if _, ok := cfg.Get("any"); ok {
		t.Error("Empty config Get() should return false")
	}

	if cfg.GetString("any", "default") != "default" {
		t.Error("Empty config should return default value")
	}

	if len(cfg.Keys()) != 0 {
		t.Error("Empty config Keys() should return empty slice")
	}

	if cfg.Has("any") {
		t.Error("Empty config Has() should return false")
	}
}

func TestConfigImpl_ConcurrentAccess(t *testing.T) {
	cfg := newConfigImplFromMap(map[string]Value{
		"key1": NewValue("value1"),
		"key2": NewValueFromInterface(42),
		"key3": NewValueFromInterface(true),
	})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 100; j++ {
				_ = cfg.GetString("key1", "")
				_ = cfg.GetInt("key2", 0)
				_ = cfg.GetBool("key3", false)
				_ = cfg.Keys()
				_ = cfg.Has("key1")
			}
		}()
	}

	wg.Wait()
}

func TestEventType_String(t *testing.T) {
	tests := []struct {
		eventType EventType
		want      string
	}{
		{EventTypeUnknown, "unknown"},
		{EventTypeCreate, "create"},
		{EventTypeUpdate, "update"},
		{EventTypeDelete, "delete"},
		{EventTypeReload, "reload"},
		{EventTypeError, "error"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.eventType.String(); got != tt.want {
				t.Errorf("EventType.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
