package postgres

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"

	"github.com/CloudRoamer/aimo-libs/config"
)

const (
	DefaultPriority = 70
	DefaultTable    = "app_config"
	DefaultKeyCol   = "key"
	DefaultValueCol = "value"
)

// Source PostgreSQL 配置源
type Source struct {
	db       *sql.DB
	table    string
	keyCol   string
	valueCol string
	priority int
}

// New 创建 PostgreSQL 配置源
func New(dsn string, opts ...Option) (*Source, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	s := &Source{
		db:       db,
		table:    DefaultTable,
		keyCol:   DefaultKeyCol,
		valueCol: DefaultValueCol,
		priority: DefaultPriority,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s, nil
}

func (s *Source) Name() string {
	return "postgres"
}

func (s *Source) Priority() int {
	return s.priority
}

func (s *Source) Load(ctx context.Context) (map[string]config.Value, error) {
	query := fmt.Sprintf(
		"SELECT %s, %s FROM %s",
		s.keyCol, s.valueCol, s.table,
	)

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query config: %w", err)
	}
	defer rows.Close()

	result := make(map[string]config.Value)
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		result[key] = config.NewValue(value)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return result, nil
}

// Watch PostgreSQL 源暂不支持监听
// 可以通过 LISTEN/NOTIFY 机制实现，但这里保持简单
func (s *Source) Watch() config.Watcher {
	return nil
}

// Close 关闭数据库连接
func (s *Source) Close() error {
	return s.db.Close()
}
