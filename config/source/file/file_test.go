package file

import (
	"context"
	"os"
	"path/filepath"
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
