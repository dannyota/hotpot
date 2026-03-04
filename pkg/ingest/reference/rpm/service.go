package rpm

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	entreference "github.com/dannyota/hotpot/pkg/storage/ent/reference"
)

const insertBatchSize = 1000

// Service handles RPM package data persistence.
type Service struct {
	client    *Client
	entClient *entreference.Client
}

// NewService creates a new RPM service.
func NewService(client *Client, entClient *entreference.Client) *Service {
	return &Service{client: client, entClient: entClient}
}

// IngestResult contains the result of an RPM package ingestion.
type IngestResult struct {
	PackageCount   int
	DurationMillis int64
}

// Ingest downloads and replaces all RPM package data.
func (s *Service) Ingest(ctx context.Context, heartbeat func(string)) (*IngestResult, error) {
	start := time.Now()

	data, err := s.client.Download(heartbeat)
	if err != nil {
		return nil, fmt.Errorf("download RPM packages: %w", err)
	}

	slog.Info("Downloaded all RPM packages", "count", len(data))
	heartbeat(fmt.Sprintf("saving %d packages", len(data)))

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete all existing entries
	deleted, err := tx.BronzeReferenceRPMPackage.Delete().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete existing RPM packages: %w", err)
	}
	slog.Info("Deleted existing RPM packages", "count", deleted)

	// Bulk insert in batches
	for i := 0; i < len(data); i += insertBatchSize {
		end := min(i+insertBatchSize, len(data))
		batch := data[i:end]

		builders := make([]*entreference.BronzeReferenceRPMPackageCreate, len(batch))
		for j, d := range batch {
			id := d.Repo + ":" + d.PackageName + ":" + d.Arch
			b := tx.BronzeReferenceRPMPackage.Create().
				SetID(id).
				SetPackageName(d.PackageName).
				SetRepo(d.Repo).
				SetArch(d.Arch).
				SetCollectedAt(now).
				SetFirstCollectedAt(now)

			if d.Version != "" {
				b.SetVersion(d.Version)
			}
			if d.RPMGroup != "" {
				b.SetRpmGroup(d.RPMGroup)
			}
			if d.Summary != "" {
				b.SetSummary(d.Summary)
			}
			if d.URL != "" {
				b.SetURL(d.URL)
			}

			builders[j] = b
		}

		if err := tx.BronzeReferenceRPMPackage.CreateBulk(builders...).Exec(ctx); err != nil {
			return nil, fmt.Errorf("bulk insert RPM packages batch %d: %w", i/insertBatchSize, err)
		}

		heartbeat(fmt.Sprintf("saved %d/%d packages", end, len(data)))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx: %w", err)
	}

	return &IngestResult{
		PackageCount:   len(data),
		DurationMillis: time.Since(start).Milliseconds(),
	}, nil
}
