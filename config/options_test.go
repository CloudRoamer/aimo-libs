package config

import "testing"

type mockMerger struct{}

func (m *mockMerger) Merge(maps ...map[string]Value) map[string]Value {
	result := make(map[string]Value)
	for _, configMap := range maps {
		for k, v := range configMap {
			result[k] = v
		}
	}
	return result
}

func TestWithMerger(t *testing.T) {
	custom := &mockMerger{}
	manager := NewManager(WithMerger(custom))

	if manager.merger != custom {
		t.Fatal("WithMerger should override default merger")
	}
}
