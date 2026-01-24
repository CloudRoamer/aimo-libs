package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/CloudRoamer/aimo-libs/config"
	"github.com/CloudRoamer/aimo-libs/config/source/env"
	"github.com/CloudRoamer/aimo-libs/config/source/file"
)

func main() {
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "./config.yaml"
	}

	fileSource, err := file.New(configFile)
	if err != nil {
		log.Fatalf("创建文件配置源失败: %v", err)
	}

	manager := config.NewManager()
	manager.AddSource(fileSource)
	manager.AddSource(env.New(env.WithPrefix("APP_")))

	if err := manager.Load(context.Background()); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	cfg := manager.Config()

	fmt.Println("=== WASM 配置加载成功 ===")
	fmt.Printf("应用名称: %s\n", cfg.GetString("app.name", "unknown"))
	fmt.Printf("应用版本: %s\n", cfg.GetString("app.version", "unknown"))
	fmt.Printf("服务端口: %d\n", cfg.GetInt("server.port", 8080))
	fmt.Printf("服务主机: %s\n", cfg.GetString("server.host", "localhost"))
	fmt.Printf("调试模式: %t\n", cfg.GetBool("debug", false))
	fmt.Printf("超时时间: %s\n", cfg.GetDuration("server.timeout", 30*time.Second))

	fmt.Println("\n=== 所有配置项 ===")
	keys := cfg.Keys()
	fmt.Printf("共有 %d 个配置项:\n", len(keys))
	for i, key := range keys {
		if val, ok := cfg.Get(key); ok {
			fmt.Printf("%d. %s = %s\n", i+1, key, val.String())
		}
	}
}
