package test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CloudRoamer/aimo-libs/config"
	"github.com/CloudRoamer/aimo-libs/config/source/env"
	"github.com/CloudRoamer/aimo-libs/config/source/file"
)

func TestManager_Integration(t *testing.T) {
	// 准备测试环境变量
	os.Setenv("TEST_DATABASE_HOST", "env.host")
	os.Setenv("TEST_SERVER_PORT", "9090")
	defer func() {
		os.Unsetenv("TEST_DATABASE_HOST")
		os.Unsetenv("TEST_SERVER_PORT")
	}()

	// 准备测试配置文件
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.json")
	jsonContent := `{
		"database": {
			"host": "file.host",
			"port": 5432
		},
		"server": {
			"port": 8080,
			"timeout": "30s"
		}
	}`
	if err := os.WriteFile(configPath, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	// 创建配置源
	fileSource, err := file.New(configPath)
	if err != nil {
		t.Fatalf("Failed to create file source: %v", err)
	}

	envSource := env.New(env.WithPrefix("TEST_"))

	// 创建管理器
	mgr := config.NewManager()
	mgr.AddSource(fileSource, envSource) // 环境变量优先级更高

	// 加载配置
	ctx := context.Background()
	if err := mgr.Load(ctx); err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	cfg := mgr.Config()

	// 测试优先级：环境变量应覆盖文件配置
	if host := cfg.GetString("database.host", ""); host != "env.host" {
		t.Errorf("Expected database.host=env.host (from env), got %s", host)
	}

	// 测试文件独有的配置
	if port := cfg.GetInt("database.port", 0); port != 5432 {
		t.Errorf("Expected database.port=5432, got %d", port)
	}

	// 测试环境变量覆盖
	if serverPort := cfg.GetInt("server.port", 0); serverPort != 9090 {
		t.Errorf("Expected server.port=9090 (from env), got %d", serverPort)
	}

	// 测试 Duration 类型
	if timeout := cfg.GetDuration("server.timeout", 0); timeout != 30*time.Second {
		t.Errorf("Expected server.timeout=30s, got %v", timeout)
	}

	// 测试 Keys
	keys := cfg.Keys()
	if len(keys) == 0 {
		t.Errorf("Expected non-empty keys")
	}

	// 测试 Has
	if !cfg.Has("database.host") {
		t.Errorf("Expected database.host to exist")
	}

	if cfg.Has("nonexistent.key") {
		t.Errorf("Expected nonexistent.key to not exist")
	}

	defer mgr.Close()
}

func TestManager_MultipleLoads(t *testing.T) {
	os.Setenv("TEST_VAR", "value1")
	defer os.Unsetenv("TEST_VAR")

	envSource := env.New(env.WithPrefix("TEST_"))
	mgr := config.NewManager()
	mgr.AddSource(envSource)

	ctx := context.Background()
	if err := mgr.Load(ctx); err != nil {
		t.Fatalf("First load failed: %v", err)
	}

	cfg := mgr.Config()
	if val := cfg.GetString("var", ""); val != "value1" {
		t.Errorf("Expected var=value1, got %s", val)
	}

	// 修改环境变量并重新加载
	os.Setenv("TEST_VAR", "value2")
	if err := mgr.Load(ctx); err != nil {
		t.Fatalf("Second load failed: %v", err)
	}

	cfg = mgr.Config()
	if val := cfg.GetString("var", ""); val != "value2" {
		t.Errorf("Expected var=value2, got %s", val)
	}

	defer mgr.Close()
}
