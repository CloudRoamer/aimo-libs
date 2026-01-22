package config

// ManagerOption 配置管理器选项
type ManagerOption func(*Manager)

// WithMerger 设置自定义合并器
func WithMerger(merger Merger) ManagerOption {
	return func(m *Manager) {
		m.merger = merger
	}
}

// Merger 定义配置合并策略接口
type Merger interface {
	// Merge 合并多个配置映射
	// 按顺序合并，后面的覆盖前面的
	Merge(maps ...map[string]Value) map[string]Value
}
