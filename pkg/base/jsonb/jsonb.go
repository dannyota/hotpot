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

// ---------------------------------------------------------------------------
// JSON query expressions (PostgreSQL)
// ---------------------------------------------------------------------------

// JSONQueryExpression builds JSON query clauses for PostgreSQL.
type JSONQueryExpression struct {
	column      string
	keys        []string
	hasKeys     bool
	equals      bool
	likes       bool
	equalsValue interface{}
	extract     bool
	path        string
}

// JSONQuery creates a new JSON query expression for the given column.
func JSONQuery(column string) *JSONQueryExpression {
	return &JSONQueryExpression{column: column}
}

// Extract extracts a JSON value at the given path.
//
//	// SELECT json_extract_path_text(attrs::json, 'name') FROM ...
//	db.Where("? = ?", jsonb.JSONQuery("attrs").Extract("name"), "foo")
func (jq *JSONQueryExpression) Extract(path string) *JSONQueryExpression {
	jq.extract = true
	jq.path = path
	return jq
}

// HasKey checks whether the JSON object contains the given nested key path.
//
//	// SELECT ... WHERE attrs::jsonb -> 'network' ? 'vpc_id'
//	db.Where(jsonb.JSONQuery("attrs").HasKey("network", "vpc_id"))
func (jq *JSONQueryExpression) HasKey(keys ...string) *JSONQueryExpression {
	jq.keys = keys
	jq.hasKeys = true
	return jq
}

// Equals checks whether the JSON value at the given key path equals value.
//
//	// SELECT ... WHERE json_extract_path_text(attrs::json, 'status') = 'RUNNING'
//	db.Where(jsonb.JSONQuery("attrs").Equals("RUNNING", "status"))
func (jq *JSONQueryExpression) Equals(value interface{}, keys ...string) *JSONQueryExpression {
	jq.keys = keys
	jq.equals = true
	jq.equalsValue = value
	return jq
}

// Likes checks whether the JSON value at the given key path matches a LIKE pattern.
//
//	// SELECT ... WHERE json_extract_path_text(attrs::json, 'name') LIKE 'prod-%'
//	db.Where(jsonb.JSONQuery("attrs").Likes("prod-%", "name"))
func (jq *JSONQueryExpression) Likes(value interface{}, keys ...string) *JSONQueryExpression {
	jq.keys = keys
	jq.likes = true
	jq.equalsValue = value
	return jq
}

// Build implements clause.Expression for PostgreSQL.
func (jq *JSONQueryExpression) Build(builder clause.Builder) {
	stmt, ok := builder.(*gorm.Statement)
	if !ok {
		return
	}

	switch {
	case jq.extract:
		builder.WriteString(fmt.Sprintf("json_extract_path_text(%v::json,", stmt.Quote(jq.column)))
		stmt.AddVar(builder, jq.path)
		builder.WriteByte(')')

	case jq.hasKeys:
		if len(jq.keys) > 0 {
			stmt.WriteQuoted(jq.column)
			stmt.WriteString("::jsonb")
			for _, key := range jq.keys[:len(jq.keys)-1] {
				stmt.WriteString(" -> ")
				stmt.AddVar(builder, key)
			}
			stmt.WriteString(" ? ")
			stmt.AddVar(builder, jq.keys[len(jq.keys)-1])
		}

	case jq.equals:
		if len(jq.keys) > 0 {
			builder.WriteString(fmt.Sprintf("json_extract_path_text(%v::json,", stmt.Quote(jq.column)))
			for idx, key := range jq.keys {
				if idx > 0 {
					builder.WriteByte(',')
				}
				stmt.AddVar(builder, key)
			}
			builder.WriteString(") = ")
			if _, ok := jq.equalsValue.(string); ok {
				stmt.AddVar(builder, jq.equalsValue)
			} else {
				stmt.AddVar(builder, fmt.Sprint(jq.equalsValue))
			}
		}

	case jq.likes:
		if len(jq.keys) > 0 {
			builder.WriteString(fmt.Sprintf("json_extract_path_text(%v::json,", stmt.Quote(jq.column)))
			for idx, key := range jq.keys {
				if idx > 0 {
					builder.WriteByte(',')
				}
				stmt.AddVar(builder, key)
			}
			builder.WriteString(") LIKE ")
			if _, ok := jq.equalsValue.(string); ok {
				stmt.AddVar(builder, jq.equalsValue)
			} else {
				stmt.AddVar(builder, fmt.Sprint(jq.equalsValue))
			}
		}
	}
}

// ---------------------------------------------------------------------------
// JSON set expression (PostgreSQL)
// ---------------------------------------------------------------------------

// JSONSetExpression builds JSONB_SET clauses for updating JSON paths.
type JSONSetExpression struct {
	column     string
	path2value map[string]interface{}
}

// JSONSet creates a new JSON set expression for the given column.
//
//	// UPDATE ... SET attrs = JSONB_SET(attrs, '{orgs,orga}', '"bar"')
//	db.UpdateColumn("attrs", jsonb.JSONSet("attrs").Set("{orgs,orga}", "bar"))
func JSONSet(column string) *JSONSetExpression {
	return &JSONSetExpression{column: column, path2value: make(map[string]interface{})}
}

// Set adds a path-value pair to the JSONB_SET expression.
// Path uses PostgreSQL syntax: {key}, {nested,key}, {array,0}.
func (js *JSONSetExpression) Set(path string, value interface{}) *JSONSetExpression {
	js.path2value[path] = value
	return js
}

// Build implements clause.Expression for PostgreSQL.
func (js *JSONSetExpression) Build(builder clause.Builder) {
	stmt, ok := builder.(*gorm.Statement)
	if !ok {
		return
	}

	var expr clause.Expression = columnExpression(js.column)
	for path, value := range js.path2value {
		if _, ok := value.(clause.Expression); ok {
			expr = gorm.Expr("JSONB_SET(?,?,?)", expr, path, value)
		} else {
			b, _ := json.Marshal(value)
			expr = gorm.Expr("JSONB_SET(?,?,?)", expr, path, string(b))
		}
	}
	stmt.AddVar(builder, expr)
}

// ---------------------------------------------------------------------------
// JSON array expression (PostgreSQL)
// ---------------------------------------------------------------------------

// JSONArrayExpression builds JSON array query clauses for PostgreSQL.
type JSONArrayExpression struct {
	contains    bool
	column      string
	equalsValue interface{}
}

// JSONArrayQuery creates a new JSON array expression for the given column.
func JSONArrayQuery(column string) *JSONArrayExpression {
	return &JSONArrayExpression{column: column}
}

// Contains checks if the JSONB array column contains the given value.
//
//	// SELECT ... WHERE tags ? 'production'
//	db.Where(jsonb.JSONArrayQuery("tags").Contains("production"))
func (ja *JSONArrayExpression) Contains(value interface{}) *JSONArrayExpression {
	ja.contains = true
	ja.equalsValue = value
	return ja
}

// Build implements clause.Expression for PostgreSQL.
func (ja *JSONArrayExpression) Build(builder clause.Builder) {
	stmt, ok := builder.(*gorm.Statement)
	if !ok {
		return
	}

	if ja.contains {
		builder.WriteString(stmt.Quote(ja.column))
		builder.WriteString(" ? ")
		builder.AddVar(stmt, ja.equalsValue)
	}
}

// ---------------------------------------------------------------------------
// Column expression helper
// ---------------------------------------------------------------------------

type columnExpression string

// Column creates a column expression that emits a properly quoted column name.
func Column(col string) columnExpression {
	return columnExpression(col)
}

// Build implements clause.Expression.
func (col columnExpression) Build(builder clause.Builder) {
	if stmt, ok := builder.(*gorm.Statement); ok {
		builder.WriteString(stmt.Quote(string(col)))
	}
}
