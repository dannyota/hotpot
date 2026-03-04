package ubuntu

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	entreference "github.com/dannyota/hotpot/pkg/storage/ent/reference"
)

const insertBatchSize = 1000

// Service handles Ubuntu package data persistence.
type Service struct {
	client    *Client
	entClient *entreference.Client
}

// NewService creates a new Ubuntu service.
func NewService(client *Client, entClient *entreference.Client) *Service {
	return &Service{client: client, entClient: entClient}
}

// IngestResult contains the result of an Ubuntu package ingestion.
type IngestResult struct {
	PackageCount   int
	DurationMillis int64
}

// Ingest downloads and replaces all Ubuntu package data.
func (s *Service) Ingest(ctx context.Context, heartbeat func(string)) (*IngestResult, error) {
	start := time.Now()

	data, err := s.client.Download(heartbeat)
	if err != nil {
		return nil, fmt.Errorf("download Ubuntu packages: %w", err)
	}

	slog.Info("Downloaded Ubuntu packages", "count", len(data))
	heartbeat(fmt.Sprintf("saving %d packages", len(data)))

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete all existing entries
	deleted, err := tx.BronzeReferenceUbuntuPackage.Delete().Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete existing Ubuntu packages: %w", err)
	}
	slog.Info("Deleted existing Ubuntu packages", "count", deleted)

	// Bulk insert in batches
	for i := 0; i < len(data); i += insertBatchSize {
		end := min(i+insertBatchSize, len(data))
		batch := data[i:end]

		builders := make([]*entreference.BronzeReferenceUbuntuPackageCreate, len(batch))
		for j, d := range batch {
			id := d.Release + ":" + d.Component + ":" + d.PackageName
			b := tx.BronzeReferenceUbuntuPackage.Create().
				SetID(id).
				SetPackageName(d.PackageName).
				SetRelease(d.Release).
				SetComponent(d.Component).
				SetSection(d.Section).
				SetCollectedAt(now).
				SetFirstCollectedAt(now)

			if d.Description != "" {
				b.SetDescription(d.Description)
			}

			builders[j] = b
		}

		if err := tx.BronzeReferenceUbuntuPackage.CreateBulk(builders...).Exec(ctx); err != nil {
			return nil, fmt.Errorf("bulk insert Ubuntu packages batch %d: %w", i/insertBatchSize, err)
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
