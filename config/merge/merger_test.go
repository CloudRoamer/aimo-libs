package merge

import (
	"testing"

	"github.com/CloudRoamer/aimo-libs/config"
)

func TestDefaultMerger_Merge(t *testing.T) {
	merger := NewDefaultMerger()

	// 模拟三个配置源（优先级从低到高）
	map1 := map[string]config.Value{
		"database.host": config.NewValue("localhost"),
		"database.port": config.NewValue("5432"),
		"server.port":   config.NewValue("8080"),
	}

	map2 := map[string]config.Value{
		"database.host": config.NewValue("prod.db.com"), // 覆盖 map1
		"server.timeout": config.NewValue("30s"),
	}

	map3 := map[string]config.Value{
		"database.port": config.NewValue("3306"), // 覆盖 map1
	}

	result := merger.Merge(map1, map2, map3)

	// 验证合并结果
	tests := []struct {
		key      string
		expected string
	}{
		{"database.host", "prod.db.com"}, // map2 覆盖
		{"database.port", "3306"},        // map3 覆盖
		{"server.port", "8080"},          // map1 保留
		{"server.timeout", "30s"},        // map2 新增
	}

	for _, tt := range tests {
		if val, ok := result[tt.key]; !ok {
			t.Errorf("Key %s not found in result", tt.key)
		} else if val.String() != tt.expected {
			t.Errorf("Key %s: expected %s, got %s", tt.key, tt.expected, val.String())
		}
	}
}

func TestDefaultMerger_EmptyMaps(t *testing.T) {
	merger := NewDefaultMerger()

	result := merger.Merge()
	if len(result) != 0 {
		t.Errorf("Expected empty result, got %d items", len(result))
	}
}
