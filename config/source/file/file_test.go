package file

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSource_LoadJSON(t *testing.T) {
	// 创建临时 JSON 文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")

	jsonContent := `{
		"database": {
			"host": "localhost",
			"port": 5432
		},
		"server": {
			"timeout": "30s"
		}
	}`

	if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	source, err := New(configPath)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	ctx := context.Background()
	values, err := source.Load(ctx)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// 检查扁平化的 key
	if val, ok := values["database.host"]; !ok || val.String() != "localhost" {
		t.Errorf("Expected database.host=localhost, got %v", val.String())
	}

	if val, ok := values["database.port"]; !ok || val.Int(0) != 5432 {
		t.Errorf("Expected database.port=5432, got %v", val.Int(0))
	}

	if val, ok := values["server.timeout"]; !ok || val.String() != "30s" {
		t.Errorf("Expected server.timeout=30s, got %v", val.String())
	}
}

func TestSource_LoadYAML(t *testing.T) {
	// 创建临时 YAML 文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
database:
  host: localhost
  port: 5432
server:
  timeout: 30s
`

	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	source, err := New(configPath)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	ctx := context.Background()
	values, err := source.Load(ctx)
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// 检查扁平化的 key
	if val, ok := values["database.host"]; !ok || val.String() != "localhost" {
		t.Errorf("Expected database.host=localhost, got %v", val.String())
	}

	if val, ok := values["database.port"]; !ok || val.Int(0) != 5432 {
		t.Errorf("Expected database.port=5432, got %v", val.Int(0))
	}
}

func TestSource_FileNotFound(t *testing.T) {
	_, err := New("/nonexistent/config.json")
	if err == nil {
		t.Errorf("Expected error for nonexistent file")
	}
}

func TestSource_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.txt")

	if err := os.WriteFile(configPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := New(configPath)
	if err == nil {
		t.Errorf("Expected error for unsupported file format")
	}
}

func TestSource_EdgeCases(t *testing.T) {
	t.Run("empty_json_file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "empty.json")

		if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		source, err := New(configPath)
		if err != nil {
			t.Fatalf("Failed to create source: %v", err)
		}

		ctx := context.Background()
		values, err := source.Load(ctx)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if len(values) != 0 {
			t.Errorf("Empty JSON should result in empty config, got %d keys", len(values))
		}
	})

	t.Run("empty_yaml_file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "empty.yaml")

		if err := os.WriteFile(configPath, []byte(""), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		source, err := New(configPath)
		if err != nil {
			t.Fatalf("Failed to create source: %v", err)
		}

		ctx := context.Background()
		values, err := source.Load(ctx)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if len(values) != 0 {
			t.Errorf("Empty YAML should result in empty config")
		}
	})

	t.Run("deeply_nested_structure", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "nested.json")

		jsonContent := `{
			"level1": {
				"level2": {
					"level3": {
						"level4": {
							"value": "deep"
						}
					}
				}
			}
		}`

		if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		source, err := New(configPath)
		if err != nil {
			t.Fatalf("Failed to create source: %v", err)
		}

		ctx := context.Background()
		values, err := source.Load(ctx)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if val, ok := values["level1.level2.level3.level4.value"]; !ok || val.String() != "deep" {
			t.Error("Deeply nested value not flattened correctly")
		}
	})

	t.Run("array_values", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "array.json")

		jsonContent := `{
			"items": ["a", "b", "c"],
			"nested": {
				"list": [1, 2, 3]
			}
		}`

		if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		source, err := New(configPath)
		if err != nil {
			t.Fatalf("Failed to create source: %v", err)
		}

		ctx := context.Background()
		values, err := source.Load(ctx)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if val, ok := values["items"]; !ok {
			t.Error("Array value should be loaded")
		} else {
			slice := val.StringSlice(nil)
			if len(slice) != 3 || slice[0] != "a" {
				t.Errorf("Array not parsed correctly: %v", slice)
			}
		}
	})

	t.Run("invalid_json", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.json")

		invalidJSON := `{"broken": "json`

		if err := os.WriteFile(configPath, []byte(invalidJSON), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		source, err := New(configPath)
		if err != nil {
			t.Fatalf("Failed to create source: %v", err)
		}

		ctx := context.Background()
		_, err = source.Load(ctx)
		if err == nil {
			t.Error("Load should fail for invalid JSON")
		}
	})

	t.Run("invalid_yaml", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.yaml")

		// YAML 解析器可能容忍某些格式问题，所以使用更明确的无效 YAML
		invalidYAML := `
key: [invalid
`

		if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		source, err := New(configPath)
		if err != nil {
			t.Fatalf("Failed to create source: %v", err)
		}

		ctx := context.Background()
		_, err = source.Load(ctx)
		if err == nil {
			t.Error("Load should fail for invalid YAML")
		}
	})

	t.Run("mixed_types", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "mixed.json")

		jsonContent := `{
			"string": "text",
			"number": 42,
			"float": 3.14,
			"bool": true,
			"null": null
		}`

		if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}

		source, err := New(configPath)
		if err != nil {
			t.Fatalf("Failed to create source: %v", err)
		}

		ctx := context.Background()
		values, err := source.Load(ctx)
		if err != nil {
			t.Fatalf("Load failed: %v", err)
		}

		if val, ok := values["string"]; !ok || val.String() != "text" {
			t.Error("String type not loaded correctly")
		}
		if val, ok := values["number"]; !ok || val.Int(0) != 42 {
			t.Error("Number type not loaded correctly")
		}
		if val, ok := values["bool"]; !ok || !val.Bool(false) {
			t.Error("Bool type not loaded correctly")
		}
	})
}

func TestSource_Priority(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	source, err := New(configPath)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	if source.Priority() != DefaultPriority {
		t.Errorf("Expected default priority %d, got %d", DefaultPriority, source.Priority())
	}
}

func TestSource_Name(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	if err := os.WriteFile(configPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	source, err := New(configPath)
	if err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	// Name 应该是 "file:" + 路径
	expectedPrefix := "file:"
	name := source.Name()
	if !strings.HasPrefix(name, expectedPrefix) {
		t.Errorf("Expected name to start with %s, got %s", expectedPrefix, name)
	}
	if !strings.Contains(name, configPath) {
		t.Errorf("Expected name to contain path %s, got %s", configPath, name)
	}
}
