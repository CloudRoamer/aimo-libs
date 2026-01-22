package merge

import "github.com/CloudRoamer/aimo-libs/config"

// Merger 定义配置合并策略接口
type Merger interface {
	// Merge 合并多个配置映射
	// 按顺序合并，后面的覆盖前面的
	Merge(maps ...map[string]config.Value) map[string]config.Value
}

// DefaultMerger 默认合并器
// 简单的覆盖策略：后面的值覆盖前面的
type DefaultMerger struct{}

// NewDefaultMerger 创建默认合并器
func NewDefaultMerger() *DefaultMerger {
	return &DefaultMerger{}
}

func (m *DefaultMerger) Merge(maps ...map[string]config.Value) map[string]config.Value {
	result := make(map[string]config.Value)

	for _, configMap := range maps {
		for k, v := range configMap {
			result[k] = v
		}
	}

	return result
}
