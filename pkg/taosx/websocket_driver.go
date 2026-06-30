// Package taosx — websocket_driver.go 导出生产级 websocket driver。
//
// 解决 issue ZoneCNH/taosx#16：sqlTDengineDriver 此前仅在 test 文件里（未导出），
// 且 WriteBatch 返回 unavailable。本文件导出 NewWebSocketDriver 工厂 + 实现
// WriteBatch（通过 SQL INSERT 渲染 Batch.Points）。
//
// binance TaosWriter 调 client.WriteBatch(ctx, batch) 落库行情数据，需要此 driver。
package taosx

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	_ "github.com/taosdata/driver-go/v3/taosWS" // 注册 taosWS sql driver
)

// NewWebSocketDriver 创建基于 taosWS 的 websocket Driver（6041 端口，不需 CGO）。
// dsn 格式：{user}:{password}@tcp({host}):6041/{database}
// 调用方用 taosx.New(ctx, cfg, WithDriver(driver)) 注入。
func NewWebSocketDriver(dsn string) (Driver, error) {
	if dsn == "" {
		return nil, fmt.Errorf("taosx.NewWebSocketDriver: dsn is required")
	}
	db, err := sql.Open("taosWS", dsn)
	if err != nil {
		return nil, fmt.Errorf("taosx.NewWebSocketDriver: open: %w", err)
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(2)
	db.SetConnMaxLifetime(5 * time.Minute)
	return &webSocketDriver{db: db}, nil
}

// webSocketDriver 是 sqlTDengineDriver 的导出版本，增加了 WriteBatch 实现。
type webSocketDriver struct {
	db *sql.DB
}

func (d *webSocketDriver) Exec(ctx context.Context, stmt Statement) (ExecResult, error) {
	result, err := d.db.ExecContext(ctx, stmt.SQL, stmt.Args...)
	if err != nil {
		return ExecResult{}, err
	}
	rowsAffected, _ := result.RowsAffected()
	return ExecResult{RowsAffected: rowsAffected}, nil
}

func (d *webSocketDriver) Query(ctx context.Context, query Query) (Rows, error) {
	rows, err := d.db.QueryContext(ctx, query.SQL, query.Args...)
	if err != nil {
		return nil, err
	}
	return &sqlRows{rows: rows}, nil
}

// WriteBatch 把 Batch.Points 渲染为 TDengine INSERT SQL 并执行。
// 每个 Point 渲染为：INSERT INTO {table} USING {stable} TAGS(...) VALUES(timestamp, ...)
// 使用自动建表语法（USING stable），无需预先创建子表。
func (d *webSocketDriver) WriteBatch(ctx context.Context, batch Batch) (WriteResult, error) {
	if len(batch.Points) == 0 {
		return WriteResult{}, nil
	}
	var written int64
	var firstErr error
	for _, p := range batch.Points {
		sqlStr, err := renderPointInsert(batch.Database, p)
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if _, err := d.db.ExecContext(ctx, sqlStr); err != nil {
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		written++
	}
	return WriteResult{
		RowsWritten:   written,
		RowsAttempted: int64(len(batch.Points)),
		Partial:       written < int64(len(batch.Points)),
	}, firstErr
}

func (d *webSocketDriver) SchemalessWrite(ctx context.Context, payload SchemalessPayload) (WriteResult, error) {
	return WriteResult{}, fmt.Errorf("taosx.webSocketDriver: SchemalessWrite not implemented, use WriteBatch")
}

func (d *webSocketDriver) Health(ctx context.Context) error {
	return d.db.PingContext(ctx)
}

func (d *webSocketDriver) Close(ctx context.Context) error {
	return d.db.Close()
}

// renderPointInsert 把 Point 渲染为 TDengine INSERT SQL。
// 格式：INSERT INTO {db}.{table} USING {stable} TAGS(...) (cols) VALUES(...)
// 使用自动建表语法（USING stable），无需预先创建子表。
func renderPointInsert(database string, p Point) (string, error) {
	if p.Table == "" {
		return "", fmt.Errorf("point table is required")
	}
	if p.Timestamp.IsZero() {
		return "", fmt.Errorf("point timestamp is required")
	}
	vals := []string{fmt.Sprintf("'%s'", p.Timestamp.Format("2006-01-02 15:04:05.000"))}
	cols := []string{"ts"}
	fieldKeys := make([]string, 0, len(p.Fields))
	for k := range p.Fields {
		fieldKeys = append(fieldKeys, k)
	}
	sort.Strings(fieldKeys)
	for _, k := range fieldKeys {
		cols = append(cols, k)
		vals = append(vals, formatValue(p.Fields[k]))
	}
	tableRef := p.Table
	if database != "" {
		tableRef = database + "." + p.Table
	}
	usingClause := ""
	if p.Stable != "" {
		stableRef := p.Stable
		if database != "" {
			stableRef = database + "." + p.Stable
		}
		tagKeys := orderedTagKeys(p.Stable, p.Tags)
		tagVals := make([]string, 0, len(tagKeys))
		for _, k := range tagKeys {
			tagVals = append(tagVals, formatValue(p.Tags[k]))
		}
		usingClause = fmt.Sprintf(" USING %s TAGS (%s)", stableRef, strings.Join(tagVals, ","))
	}
	return fmt.Sprintf("INSERT INTO %s%s (%s) VALUES (%s)",
		tableRef, usingClause, strings.Join(cols, ","), strings.Join(vals, ",")), nil
}

// orderedTagKeys returns tag keys in the supertable schema order (matching
// StableSpecs column order), not alphabetical. This ensures TAGS(...)
// values map to the correct tag columns.
var knownTagOrder = map[string][]string{
	"st_tick":  {"symbol", "product_line", "source"},
	"st_trade": {"symbol", "product_line", "source"},
	"st_bar":   {"symbol", "product_line", "source", "interval"},
}

func orderedTagKeys(stable string, tags map[string]any) []string {
	if order, ok := knownTagOrder[stable]; ok {
		result := make([]string, 0, len(order))
		for _, k := range order {
			if _, exists := tags[k]; exists {
				result = append(result, k)
			}
		}
		if len(result) == len(tags) {
			return result
		}
	}
	keys := make([]string, 0, len(tags))
	for k := range tags {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// formatValue 把 Go 值格式化为 TDengine SQL 字面量。
func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("'%s'", strings.ReplaceAll(val, "'", "\\'"))
	case int, int32, int64:
		return fmt.Sprintf("%d", val)
	case float32, float64:
		return fmt.Sprintf("%g", v)
	case bool:
		if val {
			return "true"
		}
		return "false"
	default:
		return fmt.Sprintf("'%v'", v)
	}
}

// sqlRows 包装 *sql.Rows 实现 Rows 接口。
type sqlRows struct {
	rows *sql.Rows
}

func (r *sqlRows) Columns() []string {
	cols, err := r.rows.Columns()
	if err != nil {
		return nil
	}
	return cols
}

func (r *sqlRows) Next() bool {
	return r.rows.Next()
}

func (r *sqlRows) Scan(dest ...any) error {
	return r.rows.Scan(dest...)
}

func (r *sqlRows) Close() error {
	return r.rows.Close()
}

func (r *sqlRows) Err() error {
	return r.rows.Err()
}
