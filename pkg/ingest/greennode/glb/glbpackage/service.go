package glbpackage

import (
	"context"
	"fmt"
	"time"

	entglb "github.com/dannyota/hotpot/pkg/storage/ent/greennode/glb"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/glb/bronzegreennodeglbglobalpackage"
)

// Service handles GreenNode global package ingestion.
type Service struct {
	client    *Client
	entClient *entglb.Client
	history   *HistoryService
}

// NewService creates a new global package ingestion service.
func NewService(client *Client, entClient *entglb.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of package ingestion.
type IngestResult struct {
	PackageCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches global packages from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	packages, err := s.client.ListGlobalPackages(ctx)
	if err != nil {
		return nil, fmt.Errorf("list global packages: %w", err)
	}

	dataList := make([]*GLBPackageData, 0, len(packages))
	for i := range packages {
		data, err := ConvertGLBPackage(&packages[i], projectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("convert package: %w", err)
		}
		dataList = append(dataList, data)
	}

	if err := s.savePackages(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save packages: %w", err)
	}

	return &IngestResult{
		PackageCount:   len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) savePackages(ctx context.Context, packages []*GLBPackageData) error {
	if len(packages) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range packages {
		existing, err := tx.BronzeGreenNodeGLBGlobalPackage.Query().
			Where(bronzegreennodeglbglobalpackage.ID(data.ID)).
			First(ctx)
		if err != nil && !entglb.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing package %s: %w", data.Name, err)
		}

		diff := DiffGLBPackageData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeGLBGlobalPackage.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for package %s: %w", data.Name, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGreenNodeGLBGlobalPackage.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetDescriptionEn(data.DescriptionEn).
				SetEnabled(data.Enabled).
				SetBaseSku(data.BaseSku).
				SetBaseConnectionRate(data.BaseConnectionRate).
				SetBaseDomesticTrafficTotal(data.BaseDomesticTrafficTotal).
				SetBaseNonDomesticTrafficTotal(data.BaseNonDomesticTrafficTotal).
				SetConnectionSku(data.ConnectionSku).
				SetDomesticTrafficSku(data.DomesticTrafficSku).
				SetNonDomesticTrafficSku(data.NonDomesticTrafficSku).
				SetCreatedAtAPI(data.CreatedAtAPI).
				SetUpdatedAtAPI(data.UpdatedAtAPI).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.DetailJSON != nil {
				create.SetDetailJSON(data.DetailJSON)
			}
			if data.VlbPackagesJSON != nil {
				create.SetVlbPackagesJSON(data.VlbPackagesJSON)
			}

			if _, err = create.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create package %s: %w", data.Name, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for package %s: %w", data.Name, err)
			}
		} else {
			update := tx.BronzeGreenNodeGLBGlobalPackage.UpdateOneID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetDescriptionEn(data.DescriptionEn).
				SetEnabled(data.Enabled).
				SetBaseSku(data.BaseSku).
				SetBaseConnectionRate(data.BaseConnectionRate).
				SetBaseDomesticTrafficTotal(data.BaseDomesticTrafficTotal).
				SetBaseNonDomesticTrafficTotal(data.BaseNonDomesticTrafficTotal).
				SetConnectionSku(data.ConnectionSku).
				SetDomesticTrafficSku(data.DomesticTrafficSku).
				SetNonDomesticTrafficSku(data.NonDomesticTrafficSku).
				SetCreatedAtAPI(data.CreatedAtAPI).
				SetUpdatedAtAPI(data.UpdatedAtAPI).
				SetCollectedAt(data.CollectedAt)

			if data.DetailJSON != nil {
				update.SetDetailJSON(data.DetailJSON)
			}
			if data.VlbPackagesJSON != nil {
				update.SetVlbPackagesJSON(data.VlbPackagesJSON)
			}

			if _, err = update.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update package %s: %w", data.Name, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for package %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// DeleteStalePackages removes packages not collected in the latest run.
func (s *Service) DeleteStalePackages(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	stale, err := tx.BronzeGreenNodeGLBGlobalPackage.Query().
		Where(
			bronzegreennodeglbglobalpackage.ProjectID(projectID),
			bronzegreennodeglbglobalpackage.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale packages: %w", err)
	}

	for _, pkg := range stale {
		if err := s.history.CloseHistory(ctx, tx, pkg.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for package %s: %w", pkg.ID, err)
		}
		if err := tx.BronzeGreenNodeGLBGlobalPackage.DeleteOneID(pkg.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete package %s: %w", pkg.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
