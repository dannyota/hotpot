package xeol

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	entreference "github.com/dannyota/hotpot/pkg/storage/ent/reference"
)

const insertBatchSize = 1000

// Service handles xeol product data persistence.
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
	DurationMillis int64
}

// Ingest downloads and replaces all xeol product data.
func (s *Service) Ingest(ctx context.Context, heartbeat func(string)) (*IngestResult, error) {
	start := time.Now()

	data, err := s.client.Download(heartbeat)
	if err != nil {
		return nil, fmt.Errorf("download xeol data: %w", err)
	}

	slog.Info("Downloaded xeol data", "count", len(data))
	heartbeat(fmt.Sprintf("saving %d xeol products", len(data)))

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete all existing entries
	deleted, err := tx.BronzeReferenceXeolProduct.Delete().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete existing xeol products: %w", err)
	}
	slog.Info("Deleted existing xeol products", "count", deleted)

	// Bulk insert in batches
	for i := 0; i < len(data); i += insertBatchSize {
		end := min(i+insertBatchSize, len(data))
		batch := data[i:end]

		builders := make([]*entreference.BronzeReferenceXeolProductCreate, len(batch))
		for j, d := range batch {
			b := tx.BronzeReferenceXeolProduct.Create().
				SetID(d.ID).
				SetName(d.Name).
				SetEolBool(d.EOLBool).
				SetCollectedAt(now).
				SetFirstCollectedAt(now)

			if d.PURL != "" {
				b.SetPurl(d.PURL)
			}
			if d.Permalink != "" {
				b.SetPermalink(d.Permalink)
			}
			if d.EOL != nil {
				b.SetEol(*d.EOL)
			}
			if d.LatestCycle != "" {
				b.SetLatestCycle(d.LatestCycle)
			}
			if d.ReleaseDate != nil {
				b.SetReleaseDate(*d.ReleaseDate)
			}
			if d.Latest != "" {
				b.SetLatest(d.Latest)
			}

			builders[j] = b
		}

		if err := tx.BronzeReferenceXeolProduct.CreateBulk(builders...).Exec(ctx); err != nil {
			return nil, fmt.Errorf("bulk insert xeol products batch %d: %w", i/insertBatchSize, err)
		}

		heartbeat(fmt.Sprintf("saved %d/%d xeol products", end, len(data)))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &IngestResult{
		ProductCount:   len(data),
		DurationMillis: time.Since(start).Milliseconds(),
	}, nil
}
