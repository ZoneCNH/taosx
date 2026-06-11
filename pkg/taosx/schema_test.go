package taosx

import (
	"strings"
	"testing"
)

func TestRenderCreateStableStableSQL(t *testing.T) {
	sql, err := RenderCreateStable(StableSpec{
		Name: "meters",
		Columns: []ColumnSpec{
			{Name: "ts", Type: "timestamp"},
			{Name: "value", Type: "double"},
		},
		Tags: []ColumnSpec{{Name: "location", Type: "binary(32)"}},
	})
	if err != nil {
		t.Fatalf("render stable: %v", err)
	}

	want := "CREATE STABLE IF NOT EXISTS `meters` (`ts` TIMESTAMP, `value` DOUBLE) TAGS (`location` BINARY(32))"
	if sql != want {
		t.Fatalf("unexpected SQL\nwant: %s\n got: %s", want, sql)
	}
}

func TestRenderCreateStableRejectsUnsafeIdentifier(t *testing.T) {
	_, err := RenderCreateStable(StableSpec{
		Name:    "meters;DROP",
		Columns: []ColumnSpec{{Name: "ts", Type: "timestamp"}},
	})
	if !IsKind(err, ErrorKindSchema) || strings.Contains(err.Error(), "DROP TABLE") {
		t.Fatalf("expected schema validation error, got %v", err)
	}
}
