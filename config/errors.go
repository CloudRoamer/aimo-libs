package config

import (
	"errors"
	"fmt"
)

var (
	// ErrSourceNotFound 配置源未找到
	ErrSourceNotFound = errors.New("config source not found")

	// ErrKeyNotFound 配置键不存在
	ErrKeyNotFound = errors.New("config key not found")

	// ErrTypeMismatch 类型转换失败
	ErrTypeMismatch = errors.New("config value type mismatch")

	// ErrLoadFailed 加载配置失败
	ErrLoadFailed = errors.New("failed to load config")

	// ErrWatchFailed 启动监听失败
	ErrWatchFailed = errors.New("failed to start config watch")
)

// SourceError 配置源错误
type SourceError struct {
	Source string
	Err    error
}

func (e *SourceError) Error() string {
	return fmt.Sprintf("source %s: %v", e.Source, e.Err)
}

func (e *SourceError) Unwrap() error {
	return e.Err
}
