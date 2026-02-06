package jsonb

import (
	"bytes"
	"database/sql/driver"
	"fmt"
)

// JSON is a []byte type for PostgreSQL JSONB columns.
// nil → SQL NULL, non-nil → valid JSONB value.
type JSON []byte

// Scan implements sql.Scanner for reading JSONB from the database.
func (j *JSON) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}
	switch v := value.(type) {
	case []byte:
		cp := make([]byte, len(v))
		copy(cp, v)
		*j = cp
		return nil
	case string:
		*j = []byte(v)
		return nil
	default:
		return fmt.Errorf("jsonb.JSON.Scan: unsupported type %T", value)
	}
}

// Value implements driver.Valuer for writing JSONB to the database.
func (j JSON) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return []byte(j), nil
}

// Changed reports whether two JSONB values differ.
// Handles nil (SQL NULL): both nil → false, one nil → true.
func Changed(a, b JSON) bool {
	return !bytes.Equal(a, b)
}
