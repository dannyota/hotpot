package ubuntu

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"danny.vn/hotpot/pkg/storage/ent/reference/bronzereferenceubuntupackage"

	entreference "danny.vn/hotpot/pkg/storage/ent/reference"
)

const insertBatchSize = 1000

// Service handles Ubuntu package data persistence.
type Service struct {
	entClient *entreference.Client
}

// NewService creates a new Ubuntu service.
func NewService(entClient *entreference.Client) *Service {
	return &Service{entClient: entClient}
}

// IngestFeedResult contains the result of a single Ubuntu feed ingestion.
type IngestFeedResult struct {
	Release        string
	Component      string
	PackageCount   int
	DurationMillis int64
}

// IngestFeed deletes existing packages for the given release/component and inserts new data.
func (s *Service) IngestFeed(ctx context.Context, release, component string, packages []UbuntuPackageData, heartbeat func(string)) (*IngestFeedResult, error) {
	start := time.Now()
	label := release + "/" + component

	heartbeat(fmt.Sprintf("saving %d packages for %s", len(packages), label))

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete existing entries for this release/component only
	deleted, err := tx.BronzeReferenceUbuntuPackage.Delete().
		Where(
			bronzereferenceubuntupackage.Release(release),
			bronzereferenceubuntupackage.Component(component),
		).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete existing Ubuntu packages for %s: %w", label, err)
	}
	slog.Info("Deleted existing Ubuntu packages", "feed", label, "count", deleted)

	// Bulk insert in batches
	for i := 0; i < len(packages); i += insertBatchSize {
		end := min(i+insertBatchSize, len(packages))
		batch := packages[i:end]

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
			return nil, fmt.Errorf("bulk insert Ubuntu packages batch %d for %s: %w", i/insertBatchSize, label, err)
		}

		heartbeat(fmt.Sprintf("saved %d/%d packages for %s", end, len(packages), label))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx for %s: %w", label, err)
	}

	return &IngestFeedResult{
		Release:        release,
		Component:      component,
		PackageCount:   len(packages),
		DurationMillis: time.Since(start).Milliseconds(),
	}, nil
}
