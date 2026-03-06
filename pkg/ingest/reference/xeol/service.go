package xeol

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	entreference "danny.vn/hotpot/pkg/storage/ent/reference"
)

const insertBatchSize = 1000

// Service handles xeol data persistence.
type Service struct {
	client    *Client
	entClient *entreference.Client
}

// NewService creates a new xeol service.
func NewService(client *Client, entClient *entreference.Client) *Service {
	return &Service{client: client, entClient: entClient}
}

// IngestResult contains the result of a xeol ingestion.
type IngestResult struct {
	ProductCount   int
	CycleCount     int
	PurlCount      int
	VulnCount      int
	DurationMillis int64
}

// Ingest downloads and replaces all xeol data.
func (s *Service) Ingest(ctx context.Context, heartbeat func(string)) (*IngestResult, error) {
	start := time.Now()

	data, err := s.client.Download(heartbeat)
	if err != nil {
		return nil, fmt.Errorf("download xeol data: %w", err)
	}

	slog.Info("Downloaded xeol data",
		"products", len(data.Products),
		"cycles", len(data.Cycles),
		"purls", len(data.Purls),
		"vulns", len(data.Vulns),
	)
	heartbeat(fmt.Sprintf("saving %d products, %d cycles, %d purls, %d vulns",
		len(data.Products), len(data.Cycles), len(data.Purls), len(data.Vulns)))

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete children first, then parents
	deletedVulns, err := tx.BronzeReferenceXeolVuln.Delete().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete existing xeol vulns: %w", err)
	}
	slog.Info("Deleted existing xeol vulns", "count", deletedVulns)

	deletedPurls, err := tx.BronzeReferenceXeolPurl.Delete().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete existing xeol purls: %w", err)
	}
	slog.Info("Deleted existing xeol purls", "count", deletedPurls)

	deletedCycles, err := tx.BronzeReferenceXeolCycle.Delete().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete existing xeol cycles: %w", err)
	}
	slog.Info("Deleted existing xeol cycles", "count", deletedCycles)

	deletedProducts, err := tx.BronzeReferenceXeolProduct.Delete().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete existing xeol products: %w", err)
	}
	slog.Info("Deleted existing xeol products", "count", deletedProducts)

	// Insert parents first, then children

	// Bulk insert products
	for i := 0; i < len(data.Products); i += insertBatchSize {
		end := min(i+insertBatchSize, len(data.Products))
		batch := data.Products[i:end]

		builders := make([]*entreference.BronzeReferenceXeolProductCreate, len(batch))
		for j, p := range batch {
			builders[j] = tx.BronzeReferenceXeolProduct.Create().
				SetID(p.ID).
				SetName(p.Name).
				SetNillablePermalink(nilIfEmpty(p.Permalink)).
				SetCollectedAt(now).
				SetFirstCollectedAt(now)
		}

		if err := tx.BronzeReferenceXeolProduct.CreateBulk(builders...).Exec(ctx); err != nil {
			return nil, fmt.Errorf("bulk insert xeol products batch %d: %w", i/insertBatchSize, err)
		}

		heartbeat(fmt.Sprintf("saved %d/%d xeol products", end, len(data.Products)))
	}

	// Bulk insert cycles
	for i := 0; i < len(data.Cycles); i += insertBatchSize {
		end := min(i+insertBatchSize, len(data.Cycles))
		batch := data.Cycles[i:end]

		builders := make([]*entreference.BronzeReferenceXeolCycleCreate, len(batch))
		for j, c := range batch {
			b := tx.BronzeReferenceXeolCycle.Create().
				SetID(c.ID).
				SetProductID(c.ProductID).
				SetReleaseCycle(c.ReleaseCycle).
				SetEolBool(c.EOLBool).
				SetCollectedAt(now).
				SetFirstCollectedAt(now)

			if c.EOL != nil {
				b.SetEol(*c.EOL)
			}
			if c.LatestRelease != "" {
				b.SetLatestRelease(c.LatestRelease)
			}
			if c.LatestReleaseDate != nil {
				b.SetLatestReleaseDate(*c.LatestReleaseDate)
			}
			if c.ReleaseDate != nil {
				b.SetReleaseDate(*c.ReleaseDate)
			}

			builders[j] = b
		}

		if err := tx.BronzeReferenceXeolCycle.CreateBulk(builders...).Exec(ctx); err != nil {
			return nil, fmt.Errorf("bulk insert xeol cycles batch %d: %w", i/insertBatchSize, err)
		}

		heartbeat(fmt.Sprintf("saved %d/%d xeol cycles", end, len(data.Cycles)))
	}

	// Bulk insert purls
	for i := 0; i < len(data.Purls); i += insertBatchSize {
		end := min(i+insertBatchSize, len(data.Purls))
		batch := data.Purls[i:end]

		builders := make([]*entreference.BronzeReferenceXeolPurlCreate, len(batch))
		for j, p := range batch {
			builders[j] = tx.BronzeReferenceXeolPurl.Create().
				SetID(p.ID).
				SetProductID(p.ProductID).
				SetPurl(p.PURL).
				SetCollectedAt(now).
				SetFirstCollectedAt(now)
		}

		if err := tx.BronzeReferenceXeolPurl.CreateBulk(builders...).Exec(ctx); err != nil {
			return nil, fmt.Errorf("bulk insert xeol purls batch %d: %w", i/insertBatchSize, err)
		}

		heartbeat(fmt.Sprintf("saved %d/%d xeol purls", end, len(data.Purls)))
	}

	// Bulk insert vulns
	for i := 0; i < len(data.Vulns); i += insertBatchSize {
		end := min(i+insertBatchSize, len(data.Vulns))
		batch := data.Vulns[i:end]

		builders := make([]*entreference.BronzeReferenceXeolVulnCreate, len(batch))
		for j, v := range batch {
			builders[j] = tx.BronzeReferenceXeolVuln.Create().
				SetID(v.ID).
				SetProductID(v.ProductID).
				SetVersion(v.Version).
				SetIssueCount(v.IssueCount).
				SetIssues(v.Issues).
				SetCollectedAt(now).
				SetFirstCollectedAt(now)
		}

		if err := tx.BronzeReferenceXeolVuln.CreateBulk(builders...).Exec(ctx); err != nil {
			return nil, fmt.Errorf("bulk insert xeol vulns batch %d: %w", i/insertBatchSize, err)
		}

		heartbeat(fmt.Sprintf("saved %d/%d xeol vulns", end, len(data.Vulns)))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &IngestResult{
		ProductCount:   len(data.Products),
		CycleCount:     len(data.Cycles),
		PurlCount:      len(data.Purls),
		VulnCount:      len(data.Vulns),
		DurationMillis: time.Since(start).Milliseconds(),
	}, nil
}

func nilIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
