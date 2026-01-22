package postgres

// Option PostgreSQL 配置源选项
type Option func(*Source)

// WithTable 设置配置表名
func WithTable(table string) Option {
	return func(s *Source) {
		s.table = table
	}
}

// WithColumns 设置列名
func WithColumns(keyCol, valueCol string) Option {
	return func(s *Source) {
		s.keyCol = keyCol
		s.valueCol = valueCol
	}
}

// WithPriority 设置优先级
func WithPriority(p int) Option {
	return func(s *Source) {
		s.priority = p
	}
}
