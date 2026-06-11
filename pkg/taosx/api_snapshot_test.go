package taosx

import (
	"os"
	"strings"
	"testing"
)

func TestPublicAPISnapshotFile(t *testing.T) {
	content, err := os.ReadFile("../../contracts/public_api.snapshot")
	if err != nil {
		t.Fatalf("read public api snapshot: %v", err)
	}
	text := string(content)
	for _, needle := range []string{
		"func New(context.Context, Config, ...Option) (Client, error)",
		"func NewFakeClient() *FakeClient",
		"func NewFakeDriver() *FakeDriver",
		"func RenderCreateStable(StableSpec) (string, error)",
		"const SchemalessPrecisionNanosecond SchemalessPrecision = \"ns\"",
		"type Client interface",
		"type Config struct",
		"type FakeClient struct",
		"type FakeDriver struct",
		"type HealthStatus struct",
		"type ColumnSpec struct",
		"type StableSpec struct",
		"type SchemalessPayload struct",
		"Precision SchemalessPrecision",
		"type WriteResult struct",
		"RowsAttempted int64",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("public api snapshot missing %q", needle)
		}
	}
}
