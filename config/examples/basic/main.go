package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/CloudRoamer/aimo-libs/config"
	"github.com/CloudRoamer/aimo-libs/config/source/env"
	"github.com/CloudRoamer/aimo-libs/config/source/file"
)

func main() {
	// 创建配置管理器
	mgr := config.NewManager()

	// 添加配置源（按优先级从低到高）
	// 1. 文件配置（优先级 60）
	fileSource, err := file.New("config.yaml")
	if err != nil {
		log.Fatalf("Failed to create file source: %v", err)
	}

	// 2. 环境变量（优先级 100）
	envSource := env.New(
		env.WithPrefix("APP_"),
	)

	// 添加所有配置源
	mgr.AddSource(fileSource, envSource)

	// 加载配置
	ctx := context.Background()
	if err := mgr.Load(ctx); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 获取配置
	cfg := mgr.Config()

	dbHost := cfg.GetString("database.host", "localhost")
	dbPort := cfg.GetInt("database.port", 5432)
	timeout := cfg.GetDuration("server.timeout", 30*time.Second)

	fmt.Printf("Database: %s:%d\n", dbHost, dbPort)
	fmt.Printf("Timeout: %v\n", timeout)

	// 列出所有配置键
	fmt.Println("\nAll config keys:")
	for _, key := range cfg.Keys() {
		val, _ := cfg.Get(key)
		fmt.Printf("  %s = %s\n", key, val.String())
	}

	// 关闭管理器
	defer mgr.Close()
}
