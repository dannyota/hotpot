package jsonb

import (
	"bytes"
	"context"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
)

// JSON is a []byte type for PostgreSQL JSONB columns.
// nil → SQL NULL, non-nil → valid JSONB value.
type JSON json.RawMessage

// Scan implements sql.Scanner for reading JSONB from the database.
func (j *JSON) Scan(value any) error {
	if value == nil {
		*j = JSON("null")
		return nil
	}
	var b []byte
	if s, ok := value.(fmt.Stringer); ok {
		b = []byte(s.String())
	} else {
		switch v := value.(type) {
		case []byte:
			if len(v) > 0 {
				b = make([]byte, len(v))
				copy(b, v)
			}
		case string:
			b = []byte(v)
		default:
			return errors.New(fmt.Sprint("failed to unmarshal JSONB value:", value))
		}
	}
	result := json.RawMessage(b)
	*j = JSON(result)
	return nil
}

// Value implements driver.Valuer for writing JSONB to the database.
func (j JSON) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}

// MarshalJSON implements json.Marshaler.
func (j JSON) MarshalJSON() ([]byte, error) {
	return json.RawMessage(j).MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler.
func (j *JSON) UnmarshalJSON(b []byte) error {
	result := json.RawMessage{}
	err := result.UnmarshalJSON(b)
	*j = JSON(result)
	return err
}

// String returns the JSON as a string.
func (j JSON) String() string {
	return string(j)
}

// GormDataType returns the generic GORM data type.
func (JSON) GormDataType() string {
	return "json"
}

// GormDBDataType returns the database-specific type (JSONB for PostgreSQL).
func (JSON) GormDBDataType(db *gorm.DB, field *schema.Field) string {
	switch db.Dialector.Name() {
	case "postgres":
		return "JSONB"
	}
	return ""
}

// GormValue returns a clause expression for parameterized SQL generation.
func (j JSON) GormValue(ctx context.Context, db *gorm.DB) clause.Expr {
	if len(j) == 0 {
		return gorm.Expr("NULL")
	}
	data, _ := j.MarshalJSON()
	return gorm.Expr("?", string(data))
}

// Changed reports whether two JSONB values differ.
// Handles nil (SQL NULL): both nil → false, one nil → true.
func Changed(a, b JSON) bool {
	return !bytes.Equal(a, b)
}
