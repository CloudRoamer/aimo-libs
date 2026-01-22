package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/CloudRoamer/aimo-libs/config"
	"github.com/CloudRoamer/aimo-libs/config/source/consul"
	"github.com/CloudRoamer/aimo-libs/config/source/env"
	"github.com/CloudRoamer/aimo-libs/config/source/file"
)

func main() {
	// 创建配置管理器
	mgr := config.NewManager()

	// 添加文件配置源
	fileSource, err := file.New("config.yaml")
	if err != nil {
		log.Printf("文件配置源创建失败（可选）: %v", err)
	} else {
		mgr.AddSource(fileSource)
		log.Println("已添加文件配置源")
	}

	// 添加 Consul 配置源（可选）
	consulAddr := os.Getenv("CONSUL_ADDR")
	if consulAddr != "" {
		consulSource, err := consul.New(consulAddr,
			consul.WithPrefix("config/dev/myapp"),
		)
		if err != nil {
			log.Printf("Consul 配置源创建失败（可选）: %v", err)
		} else {
			mgr.AddSource(consulSource)
			log.Println("已添加 Consul 配置源")
		}
	}

	// 添加环境变量配置源
	envSource := env.New(env.WithPrefix("APP_"))
	mgr.AddSource(envSource)
	log.Println("已添加环境变量配置源")

	// 注册配置变更回调
	mgr.OnChange(func(event config.Event, oldCfg, newCfg config.Config) {
		log.Printf("配置变更事件: 来源=%s, 类型=%s", event.Source, event.Type)

		if event.Error != nil {
			log.Printf("配置变更错误: %v", event.Error)
			return
		}

		// 检查具体配置变化
		if oldCfg != nil && newCfg != nil {
			oldHost := oldCfg.GetString("database.host", "")
			newHost := newCfg.GetString("database.host", "")
			if oldHost != newHost {
				log.Printf("数据库主机变更: %s -> %s", oldHost, newHost)
			}

			oldPort := oldCfg.GetInt("database.port", 0)
			newPort := newCfg.GetInt("database.port", 0)
			if oldPort != newPort {
				log.Printf("数据库端口变更: %d -> %d", oldPort, newPort)
			}
		}

		// 输出变更的 key 列表
		if len(event.Keys) > 0 {
			log.Printf("变更的配置项: %v", event.Keys)
		}
	})

	// 加载初始配置
	ctx := context.Background()
	if err := mgr.Load(ctx); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}
	log.Println("初始配置加载成功")

	// 输出当前配置
	cfg := mgr.Config()
	displayConfig(cfg)

	// 启动配置监听
	if err := mgr.Watch(); err != nil {
		log.Fatalf("启动配置监听失败: %v", err)
	}
	log.Println("配置监听已启动，等待变更...")

	// 模拟应用运行
	go simulateApp(cfg)

	// 等待退出信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("正在关闭...")
	mgr.Close()
}

func displayConfig(cfg config.Config) {
	fmt.Println("\n========== 当前配置 ==========")
	for _, key := range cfg.Keys() {
		val, _ := cfg.Get(key)
		fmt.Printf("%s = %s\n", key, val.String())
	}
	fmt.Println("=============================\n")
}

func simulateApp(cfg config.Config) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		dbHost := cfg.GetString("database.host", "localhost")
		dbPort := cfg.GetInt("database.port", 5432)
		log.Printf("应用运行中... 当前数据库: %s:%d", dbHost, dbPort)
	}
}
