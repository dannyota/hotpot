package cpe

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	entreference "github.com/dannyota/hotpot/pkg/storage/ent/reference"
)

const insertBatchSize = 1000

// Service handles CPE data persistence.
type Service struct {
	client    *Client
	entClient *entreference.Client
}

// NewService creates a new CPE service.
func NewService(client *Client, entClient *entreference.Client) *Service {
	return &Service{client: client, entClient: entClient}
}

// IngestResult contains the result of a CPE ingestion.
type IngestResult struct {
	CPECount       int
	DurationMillis int64
}

// Ingest downloads and replaces all CPE data.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	start := time.Now()

	data, err := s.client.Download(heartbeat)
	if err != nil {
		return nil, fmt.Errorf("download CPE data: %w", err)
	}

	slog.Info("Downloaded CPE data", "count", len(data))
	heartbeat()

	now := time.Now()

	// Delete all existing rows and bulk insert new ones in a transaction.
	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete all existing CPE entries
	deleted, err := tx.BronzeReferenceCPE.Delete().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete existing CPE: %w", err)
	}
	slog.Info("Deleted existing CPE entries", "count", deleted)

	// Bulk insert in batches
	for i := 0; i < len(data); i += insertBatchSize {
		end := min(i+insertBatchSize, len(data))
		batch := data[i:end]

		builders := make([]*entreference.BronzeReferenceCPECreate, len(batch))
		for j, d := range batch {
			b := tx.BronzeReferenceCPE.Create().
				SetID(d.CPEName).
				SetPart(d.Part).
				SetCpeVendor(d.Vendor).
				SetCpeProduct(d.Product).
				SetCpeVersion(d.Version).
				SetDeprecated(d.Deprecated).
				SetCollectedAt(now).
				SetFirstCollectedAt(now)

			if d.Title != "" {
				b.SetTitle(d.Title)
			}

			builders[j] = b
		}

		if err := tx.BronzeReferenceCPE.CreateBulk(builders...).Exec(ctx); err != nil {
			return nil, fmt.Errorf("bulk insert CPE batch %d: %w", i/insertBatchSize, err)
		}

		heartbeat()
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &IngestResult{
		CPECount:       len(data),
		DurationMillis: time.Since(start).Milliseconds(),
	}, nil
}
