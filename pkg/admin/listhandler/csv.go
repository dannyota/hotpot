package listhandler

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"net/http"
	"strings"
)

// writeCSV streams SQL rows as CSV. Each row is a JSON object produced by
// row_to_json; the first row's keys become the CSV headers.
func writeCSV(w http.ResponseWriter, rows *sql.Rows) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=export.csv")

	cw := csv.NewWriter(w)
	defer cw.Flush()

	var headers []string
	for rows.Next() {
		var raw json.RawMessage
		if err := rows.Scan(&raw); err != nil {
			return
		}

		var obj map[string]json.RawMessage
		if err := json.Unmarshal(raw, &obj); err != nil {
			return
		}

		// Write headers from the first row.
		if headers == nil {
			headers = extractJSONKeys(raw)
			if err := cw.Write(headers); err != nil {
				return
			}
		}

		record := make([]string, len(headers))
		for i, h := range headers {
			record[i] = jsonValueToString(obj[h])
		}
		if err := cw.Write(record); err != nil {
			return
		}
	}
}

// extractJSONKeys returns the keys of a JSON object in their original order.
func extractJSONKeys(raw json.RawMessage) []string {
	dec := json.NewDecoder(strings.NewReader(string(raw)))
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return nil
	}
	var keys []string
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			break
		}
		if key, ok := t.(string); ok {
			keys = append(keys, key)
			var v json.RawMessage
			if err := dec.Decode(&v); err != nil {
				break
			}
		}
	}
	return keys
}

// jsonValueToString converts a JSON value to a CSV-friendly string.
func jsonValueToString(v json.RawMessage) string {
	if len(v) == 0 {
		return ""
	}
	var s string
	if err := json.Unmarshal(v, &s); err == nil {
		return s
	}
	return strings.TrimSpace(string(v))
}
