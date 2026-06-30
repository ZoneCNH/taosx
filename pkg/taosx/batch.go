package taosx

import (
	"time"
)

type Point struct {
	Table     string
	Stable    string
	Timestamp time.Time
	Tags      map[string]any
	Fields    map[string]any
}

type Batch struct {
	Database string
	Points   []Point
}

type WriteResult struct {
	RowsWritten   int64
	RowsAttempted int64
	Partial       bool
}

func (b Batch) Validate() error {
	const op = "Batch.Validate"
	if b.Database == "" {
		return validationError(op, "database is required", nil)
	}
	if len(b.Points) == 0 {
		return validationError(op, "points are required", nil)
	}
	for i, point := range b.Points {
		if err := validateIdentifier(point.Table); err != nil {
			return validationError(op, "points["+itoa(i)+"].table: "+err.Error(), err)
		}
		if point.Timestamp.IsZero() {
			return validationError(op, "points["+itoa(i)+"].timestamp is required", nil)
		}
		if len(point.Fields) == 0 {
			return validationError(op, "points["+itoa(i)+"].fields are required", nil)
		}
	}
	return nil
}
