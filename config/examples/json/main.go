package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"github.com/tetratelabs/wazero/sys"
)

func main() {
	modulePath := os.Getenv("CONFIG_WASM_PATH")
	if modulePath == "" {
		modulePath = defaultWasmPath()
	}

	configFile := os.Getenv("CONFIG_FILE")
	if configFile == "" {
		configFile = "./config.json"
	}

	fmt.Printf("加载 WASM: %s\n", modulePath)
	if err := runWasm(modulePath, configFile); err != nil {
		log.Fatalf("运行 WASM 失败: %v", err)
	}
}

func defaultWasmPath() string {
	base := filepath.Join("..", "..", "..", "dist", "config.wasm")
	if absPath, err := filepath.Abs(base); err == nil {
		return absPath
	}
	return base
}

func runWasm(modulePath string, configFile string) error {
	ctx := context.Background()
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)

	_, err := wasi_snapshot_preview1.Instantiate(ctx, runtime)
	if err != nil {
		return err
	}

	moduleBytes, err := os.ReadFile(modulePath)
	if err != nil {
		return err
	}

	moduleConfig := wazero.NewModuleConfig().
		WithStdout(os.Stdout).
		WithStderr(os.Stderr).
		WithEnv("CONFIG_FILE", configFile)

	if value := os.Getenv("APP_SERVER_PORT"); value != "" {
		moduleConfig = moduleConfig.WithEnv("APP_SERVER_PORT", value)
	}
	if value := os.Getenv("APP_DEBUG"); value != "" {
		moduleConfig = moduleConfig.WithEnv("APP_DEBUG", value)
	}

	fsConfig := wazero.NewFSConfig().WithDirMount(".", "/")
	moduleConfig = moduleConfig.WithFSConfig(fsConfig)

	compiled, err := runtime.CompileModule(ctx, moduleBytes)
	if err != nil {
		return err
	}

	_, err = runtime.InstantiateModule(ctx, compiled, moduleConfig)
	if err != nil {
		if exitErr, ok := err.(*sys.ExitError); ok && exitErr.ExitCode() == 0 {
			return nil
		}
		return err
	}

	return nil
}
