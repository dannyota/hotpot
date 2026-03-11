package gcplogging

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// BQClient queries BigQuery Log Analytics for aggregated access log data.
type BQClient struct {
	client *bigquery.Client
	table  string // "project.dataset._AllLogs"
	filter string // BigQuery WHERE clause fragment
}

// NewBQClient creates a BigQuery client for Log Analytics queries.
// If creds is non-empty, uses explicit credentials; otherwise falls back to ADC.
func NewBQClient(ctx context.Context, creds []byte, projectID, table, filter string) (*BQClient, error) {
	var opts []option.ClientOption
	if len(creds) > 0 {
		opts = append(opts, option.WithCredentialsJSON(creds))
	}

	client, err := bigquery.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, fmt.Errorf("create bigquery client: %w", err)
	}

	return &BQClient{
		client: client,
		table:  table,
		filter: filter,
	}, nil
}

// Close releases BigQuery client resources.
func (c *BQClient) Close() error {
	return c.client.Close()
}

// HttpCountRow holds a single aggregated HTTP count result.
type HttpCountRow struct {
	URI              string  `bigquery:"uri"`
	Method           string  `bigquery:"method"`
	StatusCode       int64   `bigquery:"status_code"`
	RequestCount     int64   `bigquery:"request_count"`
	TotalBodyBytes   int64   `bigquery:"total_body_bytes"`
	TotalRequestTime float64 `bigquery:"total_request_time"`
	MaxRequestTime   float64 `bigquery:"max_request_time"`
	HTTPHost         string  `bigquery:"http_host"`
}

// UserAgentRow holds a single aggregated user agent result.
type UserAgentRow struct {
	URI          string `bigquery:"uri"`
	Method       string `bigquery:"method"`
	UserAgent    string `bigquery:"user_agent"`
	RequestCount int64  `bigquery:"request_count"`
}

// ClientIPRow holds a single aggregated client IP result.
type ClientIPRow struct {
	URI          string `bigquery:"uri"`
	Method       string `bigquery:"method"`
	ClientIP     string `bigquery:"client_ip"`
	RequestCount int64  `bigquery:"request_count"`
}

// QueryHttpCounts returns aggregated HTTP counts grouped by (uri, method, status).
func (c *BQClient) QueryHttpCounts(ctx context.Context, fm map[string]string, start, end time.Time) ([]HttpCountRow, error) {
	sql := fmt.Sprintf(`SELECT
  %s AS uri,
  %s AS method,
  CAST(%s AS INT64) AS status_code,
  COUNT(*) AS request_count,
  COALESCE(SUM(CAST(%s AS INT64)), 0) AS total_body_bytes,
  COALESCE(SUM(CAST(%s AS FLOAT64)), 0) AS total_request_time,
  COALESCE(MAX(CAST(%s AS FLOAT64)), 0) AS max_request_time,
  ANY_VALUE(%s) AS http_host
FROM `+"`%s`"+`
WHERE timestamp >= @start_time AND timestamp < @end_time AND (%s)
GROUP BY uri, method, status_code`,
		jsonField(fm, "uri"),
		jsonField(fm, "method"),
		jsonField(fm, "status"),
		jsonField(fm, "body_bytes_sent"),
		jsonField(fm, "request_time"),
		jsonField(fm, "request_time"),
		jsonField(fm, "http_host"),
		c.table,
		c.filterClause(),
	)

	var rows []HttpCountRow
	if err := c.runQuery(ctx, sql, start, end, func(it *bigquery.RowIterator) error {
		for {
			var row HttpCountRow
			err := it.Next(&row)
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			rows = append(rows, row)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("query http counts: %w", err)
	}
	return rows, nil
}

// QueryUserAgents returns aggregated user agents grouped by (uri, method, user_agent).
func (c *BQClient) QueryUserAgents(ctx context.Context, fm map[string]string, start, end time.Time) ([]UserAgentRow, error) {
	sql := fmt.Sprintf(`SELECT
  %s AS uri,
  %s AS method,
  %s AS user_agent,
  COUNT(*) AS request_count
FROM `+"`%s`"+`
WHERE timestamp >= @start_time AND timestamp < @end_time AND (%s)
GROUP BY uri, method, user_agent`,
		jsonField(fm, "uri"),
		jsonField(fm, "method"),
		jsonField(fm, "http_user_agent"),
		c.table,
		c.filterClause(),
	)

	var rows []UserAgentRow
	if err := c.runQuery(ctx, sql, start, end, func(it *bigquery.RowIterator) error {
		for {
			var row UserAgentRow
			err := it.Next(&row)
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			rows = append(rows, row)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("query user agents: %w", err)
	}
	return rows, nil
}

// QueryClientIPs returns aggregated client IPs grouped by (uri, method, client_ip).
func (c *BQClient) QueryClientIPs(ctx context.Context, fm map[string]string, start, end time.Time) ([]ClientIPRow, error) {
	sql := fmt.Sprintf(`SELECT
  %s AS uri,
  %s AS method,
  %s AS client_ip,
  COUNT(*) AS request_count
FROM `+"`%s`"+`
WHERE timestamp >= @start_time AND timestamp < @end_time AND (%s)
GROUP BY uri, method, client_ip`,
		jsonField(fm, "uri"),
		jsonField(fm, "method"),
		jsonField(fm, "remote_addr"),
		c.table,
		c.filterClause(),
	)

	var rows []ClientIPRow
	if err := c.runQuery(ctx, sql, start, end, func(it *bigquery.RowIterator) error {
		for {
			var row ClientIPRow
			err := it.Next(&row)
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			rows = append(rows, row)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("query client ips: %w", err)
	}
	return rows, nil
}

// filterClause returns the BQ filter wrapped for use in a WHERE clause.
// If no filter is configured, returns "TRUE" so the SQL stays valid.
func (c *BQClient) filterClause() string {
	if c.filter == "" {
		return "TRUE"
	}
	return c.filter
}

// runQuery executes a parameterized BigQuery query with time range parameters.
func (c *BQClient) runQuery(ctx context.Context, sql string, start, end time.Time, scan func(*bigquery.RowIterator) error) error {
	q := c.client.Query(sql)
	q.Parameters = []bigquery.QueryParameter{
		{Name: "start_time", Value: start},
		{Name: "end_time", Value: end},
	}

	it, err := q.Read(ctx)
	if err != nil {
		return err
	}
	return scan(it)
}

var validFieldName = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_.]*$`)

// jsonField returns a BigQuery JSON_VALUE expression for the given field name,
// applying field_mapping if configured.
func jsonField(fm map[string]string, name string) string {
	field := name
	if fm != nil {
		if m, ok := fm[name]; ok {
			field = m
		}
	}
	if !validFieldName.MatchString(field) {
		// Fall back to canonical name if mapping is invalid.
		field = name
	}
	return "JSON_VALUE(json_payload." + field + ")"
}
