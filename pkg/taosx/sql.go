package taosx

import "strings"

type Statement struct {
	SQL  string
	Args []any
}

type Query struct {
	SQL  string
	Args []any
}

type ExecResult struct {
	RowsAffected int64
}

type Rows interface {
	Columns() []string
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close() error
}

func NewStatement(sql string, args ...any) Statement {
	return Statement{SQL: sql, Args: append([]any(nil), args...)}
}

func NewQuery(sql string, args ...any) Query {
	return Query{SQL: sql, Args: append([]any(nil), args...)}
}

func validateSQL(op string, sql string) error {
	if strings.TrimSpace(sql) == "" {
		return validationError(op, "sql is required", nil)
	}
	return nil
}
