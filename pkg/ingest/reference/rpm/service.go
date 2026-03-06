package rpm

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"danny.vn/hotpot/pkg/storage/ent/reference/bronzereferencerpmpackage"

	entreference "danny.vn/hotpot/pkg/storage/ent/reference"
)

const insertBatchSize = 1000

// Service handles RPM package data persistence.
type Service struct {
	entClient *entreference.Client
}

// NewService creates a new RPM service.
func NewService(entClient *entreference.Client) *Service {
	return &Service{entClient: entClient}
}

// IngestRepoResult contains the result of a single RPM repo ingestion.
type IngestRepoResult struct {
	RepoName       string
	PackageCount   int
	DurationMillis int64
}

// IngestRepo deletes existing packages for the given repo and inserts the new data.
func (s *Service) IngestRepo(ctx context.Context, repoName string, packages []RPMPackageData, heartbeat func(string)) (*IngestRepoResult, error) {
	start := time.Now()

	heartbeat(fmt.Sprintf("saving %d packages for %s", len(packages), repoName))

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Delete existing entries for this repo only
	deleted, err := tx.BronzeReferenceRPMPackage.Delete().
		Where(bronzereferencerpmpackage.Repo(repoName)).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("delete existing RPM packages for %s: %w", repoName, err)
	}
	slog.Info("Deleted existing RPM packages", "repo", repoName, "count", deleted)

	// Bulk insert in batches
	for i := 0; i < len(packages); i += insertBatchSize {
		end := min(i+insertBatchSize, len(packages))
		batch := packages[i:end]

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
			return nil, fmt.Errorf("bulk insert RPM packages batch %d for %s: %w", i/insertBatchSize, repoName, err)
		}

		heartbeat(fmt.Sprintf("saved %d/%d packages for %s", end, len(packages), repoName))
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit tx for %s: %w", repoName, err)
	}

	return &IngestRepoResult{
		RepoName:       repoName,
		PackageCount:   len(packages),
		DurationMillis: time.Since(start).Milliseconds(),
	}, nil
}
