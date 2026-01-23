//go:build linux || darwin

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"plugin"
	"reflect"
	"time"

	"github.com/CloudRoamer/aimo-libs/config"
)

func main() {
	// 1. 加载 Plugin
	pluginPath := os.Getenv("CONFIG_PLUGIN_PATH")
	if pluginPath == "" {
		pluginPath = "../../dist/darwin/config.so"
	}

	fmt.Printf("加载 Plugin: %s\n", pluginPath)
	p, err := plugin.Open(pluginPath)
	if err != nil {
		log.Fatalf("加载 Plugin 失败: %v", err)
	}

	// 2. 查找导出的符号
	newManagerSym, err := p.Lookup("NewManager")
	if err != nil {
		log.Fatalf("查找 NewManager 失败: %v", err)
	}

	newFileSourceSym, err := p.Lookup("NewFileSource")
	if err != nil {
		log.Fatalf("查找 NewFileSource 失败: %v", err)
	}

	newEnvSourceSym, err := p.Lookup("NewEnvSource")
	if err != nil {
		log.Fatalf("查找 NewEnvSource 失败: %v", err)
	}

	envWithPrefixSym, err := p.Lookup("EnvWithPrefix")
	if err != nil {
		log.Fatalf("查找 EnvWithPrefix 失败: %v", err)
	}

	// 3. 通过反射调用（Plugin 导出的是变量）
	// 创建配置管理器
	fmt.Println("\n=== 创建配置管理器 ===")
	newManagerVal := reflect.ValueOf(newManagerSym).Elem()
	managerResults := newManagerVal.Call([]reflect.Value{})
	manager := managerResults[0].Interface().(*config.Manager)

	// 4. 添加文件配置源
	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "./config.yaml"
	}

	fmt.Printf("加载配置文件: %s\n", configFile)
	newFileSourceVal := reflect.ValueOf(newFileSourceSym)
	// NewFileSource 是函数，不需要 Elem()
	fileSourceResults := newFileSourceVal.Call([]reflect.Value{
		reflect.ValueOf(configFile),
	})

	if !fileSourceResults[1].IsNil() {
		err := fileSourceResults[1].Interface().(error)
		log.Fatalf("创建文件配置源失败: %v", err)
	}
	fileSource := fileSourceResults[0].Interface().(config.Source)

	// 5. 添加环境变量配置源（前缀为 APP_）
	fmt.Println("创建环境变量配置源（前缀: APP_）")

	// 先创建 EnvWithPrefix 选项
	envWithPrefixVal := reflect.ValueOf(envWithPrefixSym).Elem()
	prefixResults := envWithPrefixVal.Call([]reflect.Value{
		reflect.ValueOf("APP_"),
	})
	prefixOption := prefixResults[0]

	// 创建 Env Source
	newEnvSourceVal := reflect.ValueOf(newEnvSourceSym)
	envSourceResults := newEnvSourceVal.Call([]reflect.Value{prefixOption})
	envSource := envSourceResults[0].Interface().(config.Source)

	// 6. 添加配置源到管理器
	manager.AddSource(fileSource)
	manager.AddSource(envSource)

	// 7. 加载配置
	ctx := context.Background()
	if err := manager.Load(ctx); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	// 8. 读取配置值
	cfg := manager.Config()

	fmt.Println("\n=== 配置已加载 ===")
	fmt.Printf("应用名称: %s\n", cfg.GetString("app.name", "unknown"))
	fmt.Printf("应用版本: %s\n", cfg.GetString("app.version", "unknown"))
	fmt.Printf("服务端口: %d\n", cfg.GetInt("server.port", 8080))
	fmt.Printf("服务主机: %s\n", cfg.GetString("server.host", "localhost"))
	fmt.Printf("调试模式: %t\n", cfg.GetBool("debug", false))
	fmt.Printf("超时时间: %s\n", cfg.GetDuration("server.timeout", 30*time.Second))

	// 9. 列出所有配置键
	fmt.Println("\n=== 所有配置项 ===")
	keys := cfg.Keys()
	fmt.Printf("共有 %d 个配置项:\n", len(keys))
	for i, key := range keys {
		if val, ok := cfg.Get(key); ok {
			fmt.Printf("%d. %s = %s\n", i+1, key, val.String())
		}
	}

	// 10. 启动配置监听（可选）
	fmt.Println("\n=== 启动配置监听器 ===")
	manager.OnChange(func(event config.Event, oldConfig, newConfig config.Config) {
		fmt.Printf("\n[配置变更] 类型=%s, 来源=%s, 时间=%s\n",
			event.Type, event.Source, event.Timestamp.Format("15:04:05"))
		if len(event.Keys) > 0 {
			fmt.Printf("变更的键: %v\n", event.Keys)
		}
		if event.Error != nil {
			fmt.Printf("错误: %v\n", event.Error)
		}
	})

	if err := manager.Watch(); err != nil {
		log.Printf("启动监听器失败: %v", err)
	}

	// 11. 等待一段时间观察配置变更
	fmt.Println("\n监听配置变更中... (5秒后自动退出，或按 Ctrl+C)")
	fmt.Println("尝试修改 config.yaml 文件，观察热更新效果")
	fmt.Println("或设置环境变量: export APP_DEBUG=true && go run main.go")

	time.Sleep(5 * time.Second)

	// 12. 清理
	fmt.Println("\n正在关闭配置管理器...")
	if err := manager.Close(); err != nil {
		log.Printf("关闭管理器失败: %v", err)
	}

	fmt.Println("程序退出")
}
