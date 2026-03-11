package apicatalog

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"log/slog"

	"danny.vn/hotpot/pkg/base/config"
	entapicatalog "danny.vn/hotpot/pkg/storage/ent/apicatalog"
)

// Activities holds dependencies for API catalog Temporal activities.
type Activities struct {
	configService *config.Service
	entClient     *entapicatalog.Client
}

// NewActivities creates an Activities instance.
func NewActivities(configService *config.Service, entClient *entapicatalog.Client) *Activities {
	return &Activities{
		configService: configService,
		entClient:     entClient,
	}
}

// ImportCSVActivity function reference for Temporal registration.
var ImportCSVActivity = (*Activities).ImportCSV

// ImportCSVParams holds input for the ImportCSV activity.
type ImportCSVParams struct {
	FilePath    string
	CSVData     []byte
	LogSourceID string
	SourceFile  string
}

// ImportCSVResult holds output from the ImportCSV activity.
type ImportCSVResult struct {
	Created int
	Updated int
}

// ImportCSV imports API endpoint definitions from a CSV file into bronze.
func (a *Activities) ImportCSV(ctx context.Context, params ImportCSVParams) (*ImportCSVResult, error) {
	logger := slog.Default()

	var reader io.Reader
	if len(params.CSVData) > 0 {
		reader = bytes.NewReader(params.CSVData)
	} else if params.FilePath != "" {
		f, err := os.Open(params.FilePath)
		if err != nil {
			return nil, fmt.Errorf("open csv file: %w", err)
		}
		defer f.Close()
		reader = f
	} else {
		return nil, fmt.Errorf("either FilePath or CSVData must be provided")
	}

	records, err := parseCSV(reader)
	if err != nil {
		return nil, fmt.Errorf("parse csv: %w", err)
	}

	sourceFile := params.SourceFile
	if sourceFile == "" && params.FilePath != "" {
		// Use just the filename, not the full path.
		parts := strings.Split(params.FilePath, "/")
		sourceFile = parts[len(parts)-1]
	}

	now := time.Now()
	var created, updated int

	// Load existing rows keyed by (uri, method, route_status) for dedup.
	existingByKey := make(map[string]string) // dedupKey → resource_id

	allRows, err := a.entClient.BronzeApicatalogEndpointsRaw.Query().
		Select("name", "upstream", "uri", "method", "route_status").
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query existing endpoints: %w", err)
	}
	for _, row := range allRows {
		key := row.Name + "||" + row.Upstream + "||" + row.URI + "||" + row.Method + "||" + row.RouteStatus
		existingByKey[key] = row.ID
	}

	for _, rec := range records {
		dedupKey := rec.name + "||" + rec.upstream + "||" + rec.uri + "||" + rec.method + "||" + rec.routeStatus

		if existingID, ok := existingByKey[dedupKey]; ok {
			// Update existing row.
			_, err := a.entClient.BronzeApicatalogEndpointsRaw.UpdateOneID(existingID).
				SetCollectedAt(now).
				SetNillableLogSourceID(nilIfEmpty(params.LogSourceID)).
				SetNillableName(nilIfEmpty(rec.name)).
				SetNillableServiceName(nilIfEmpty(rec.serviceName)).
				SetNillableUpstream(nilIfEmpty(rec.upstream)).
				SetURI(rec.uri).
				SetMethod(rec.method).
				SetRouteStatus(rec.routeStatus).
				SetNillablePluginAuth(nilIfEmpty(rec.pluginAuth)).
				SetNillablePluginAuthEnable(nilIfEmpty(rec.pluginAuthEnable)).
				SetNillableSourceFile(nilIfEmpty(sourceFile)).
				Save(ctx)
			if err != nil {
				return nil, fmt.Errorf("update endpoint %s: %w", existingID, err)
			}
			updated++
		} else {
			// Create new row with UUID.
			resourceID := uuid.New().String()
			create := a.entClient.BronzeApicatalogEndpointsRaw.Create().
				SetID(resourceID).
				SetCollectedAt(now).
				SetFirstCollectedAt(now).
				SetURI(rec.uri).
				SetMethod(rec.method).
				SetRouteStatus(rec.routeStatus)

			if params.LogSourceID != "" {
				create.SetLogSourceID(params.LogSourceID)
			}
			if rec.name != "" {
				create.SetName(rec.name)
			}
			if rec.serviceName != "" {
				create.SetServiceName(rec.serviceName)
			}
			if rec.upstream != "" {
				create.SetUpstream(rec.upstream)
			}
			if rec.pluginAuth != "" {
				create.SetPluginAuth(rec.pluginAuth)
			}
			if rec.pluginAuthEnable != "" {
				create.SetPluginAuthEnable(rec.pluginAuthEnable)
			}
			if sourceFile != "" {
				create.SetSourceFile(sourceFile)
			}

			if err := create.Exec(ctx); err != nil {
				return nil, fmt.Errorf("create endpoint %s: %w", resourceID, err)
			}
			existingByKey[dedupKey] = resourceID
			created++
		}
	}

	logger.InfoContext(ctx, "CSV import complete",
		"created", created,
		"updated", updated,
		"total", len(records))

	return &ImportCSVResult{
		Created: created,
		Updated: updated,
	}, nil
}

type csvRecord struct {
	name             string
	serviceName      string
	upstream         string
	uri              string
	method           string
	routeStatus      string
	pluginAuth       string
	pluginAuthEnable string
}

// normalizeHeader converts a CSV header to a canonical key.
// Lowercases, trims, replaces spaces with underscores, and applies aliases
// so both "service name" and "service_name" work, "method" and "methods" work, etc.
func normalizeHeader(h string) string {
	key := strings.ToLower(strings.TrimSpace(h))
	key = strings.ReplaceAll(key, " ", "_")

	// Aliases: map variant names to canonical keys.
	aliases := map[string]string{
		"method":  "methods",
		"status":  "route_status",
	}
	if canonical, ok := aliases[key]; ok {
		return canonical
	}
	return key
}

// parseCSV reads a CSV file and returns parsed records.
// Columns are matched by header name (case-insensitive, spaces treated as underscores).
// Recognized columns: name, service_name, upstream, uri, methods (or method),
// route_status (or status), plugin_auth, plugin_auth_enable.
func parseCSV(reader io.Reader) ([]csvRecord, error) {
	csvReader := csv.NewReader(reader)
	csvReader.LazyQuotes = true
	csvReader.TrimLeadingSpace = true

	// Read header row.
	header, err := csvReader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}

	// Build column index from normalized header.
	colIdx := make(map[string]int)
	for i, h := range header {
		colIdx[normalizeHeader(h)] = i
	}

	// Require at least uri column.
	if _, ok := colIdx["uri"]; !ok {
		return nil, fmt.Errorf("CSV missing required 'uri' column; found headers: %v", header)
	}

	var records []csvRecord
	for {
		row, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}

		getCol := func(name string) string {
			if idx, ok := colIdx[name]; ok && idx < len(row) {
				return strings.TrimSpace(row[idx])
			}
			return ""
		}

		uri := getCol("uri")
		if uri == "" {
			continue
		}

		records = append(records, csvRecord{
			name:             getCol("name"),
			serviceName:      getCol("service_name"),
			upstream:         getCol("upstream"),
			uri:              uri,
			method:           getCol("methods"),
			routeStatus:      getCol("route_status"),
			pluginAuth:       getCol("plugin_auth"),
			pluginAuthEnable: getCol("plugin_auth_enable"),
		})
	}

	return records, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
