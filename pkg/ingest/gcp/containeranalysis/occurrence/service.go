package occurrence

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcontaineranalysisoccurrence"
)

// Service handles Grafeas occurrence ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Grafeas occurrence ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of occurrence ingestion.
type IngestResult struct {
	ProjectID       string
	OccurrenceCount int
	CollectedAt     time.Time
	DurationMillis  int64
}

// Ingest fetches Grafeas occurrences from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	rawOccurrences, err := s.client.ListOccurrences(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list occurrences: %w", err)
	}

	occurrenceDataList := make([]*OccurrenceData, 0, len(rawOccurrences))
	for _, raw := range rawOccurrences {
		data := ConvertOccurrence(raw, projectID, collectedAt)
		occurrenceDataList = append(occurrenceDataList, data)
	}

	if err := s.saveOccurrences(ctx, occurrenceDataList); err != nil {
		return nil, fmt.Errorf("failed to save occurrences: %w", err)
	}

	return &IngestResult{
		ProjectID:       projectID,
		OccurrenceCount: len(occurrenceDataList),
		CollectedAt:     collectedAt,
		DurationMillis:  time.Since(startTime).Milliseconds(),
	}, nil
}

// saveOccurrences saves Grafeas occurrences to the database with history tracking.
func (s *Service) saveOccurrences(ctx context.Context, occurrences []*OccurrenceData) error {
	if len(occurrences) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, occData := range occurrences {
		// Load existing occurrence
		existing, err := tx.BronzeGCPContainerAnalysisOccurrence.Query().
			Where(bronzegcpcontaineranalysisoccurrence.ID(occData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing occurrence %s: %w", occData.ID, err)
		}

		// Compute diff
		diff := DiffOccurrenceData(existing, occData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPContainerAnalysisOccurrence.UpdateOneID(occData.ID).
				SetCollectedAt(occData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for occurrence %s: %w", occData.ID, err)
			}
			continue
		}

		// Create or update occurrence
		if existing == nil {
			create := tx.BronzeGCPContainerAnalysisOccurrence.Create().
				SetID(occData.ID).
				SetProjectID(occData.ProjectID).
				SetCollectedAt(occData.CollectedAt).
				SetFirstCollectedAt(occData.CollectedAt).
				SetKind(occData.Kind)

			if occData.ResourceURI != "" {
				create.SetResourceURI(occData.ResourceURI)
			}
			if occData.NoteName != "" {
				create.SetNoteName(occData.NoteName)
			}
			if occData.Remediation != "" {
				create.SetRemediation(occData.Remediation)
			}
			if occData.CreateTime != "" {
				create.SetCreateTime(occData.CreateTime)
			}
			if occData.UpdateTime != "" {
				create.SetUpdateTime(occData.UpdateTime)
			}
			if occData.VulnerabilityJSON != nil {
				create.SetVulnerabilityJSON(occData.VulnerabilityJSON)
			}
			if occData.BuildJSON != nil {
				create.SetBuildJSON(occData.BuildJSON)
			}
			if occData.ImageJSON != nil {
				create.SetImageJSON(occData.ImageJSON)
			}
			if occData.PackageJSON != nil {
				create.SetPackageJSON(occData.PackageJSON)
			}
			if occData.DeploymentJSON != nil {
				create.SetDeploymentJSON(occData.DeploymentJSON)
			}
			if occData.DiscoveryJSON != nil {
				create.SetDiscoveryJSON(occData.DiscoveryJSON)
			}
			if occData.AttestationJSON != nil {
				create.SetAttestationJSON(occData.AttestationJSON)
			}
			if occData.UpgradeJSON != nil {
				create.SetUpgradeJSON(occData.UpgradeJSON)
			}
			if occData.ComplianceJSON != nil {
				create.SetComplianceJSON(occData.ComplianceJSON)
			}
			if occData.DsseAttestationJSON != nil {
				create.SetDsseAttestationJSON(occData.DsseAttestationJSON)
			}
			if occData.SbomReferenceJSON != nil {
				create.SetSbomReferenceJSON(occData.SbomReferenceJSON)
			}
			if occData.EnvelopeJSON != nil {
				create.SetEnvelopeJSON(occData.EnvelopeJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create occurrence %s: %w", occData.ID, err)
			}
		} else {
			update := tx.BronzeGCPContainerAnalysisOccurrence.UpdateOneID(occData.ID).
				SetProjectID(occData.ProjectID).
				SetCollectedAt(occData.CollectedAt).
				SetKind(occData.Kind)

			if occData.ResourceURI != "" {
				update.SetResourceURI(occData.ResourceURI)
			}
			if occData.NoteName != "" {
				update.SetNoteName(occData.NoteName)
			}
			if occData.Remediation != "" {
				update.SetRemediation(occData.Remediation)
			}
			if occData.CreateTime != "" {
				update.SetCreateTime(occData.CreateTime)
			}
			if occData.UpdateTime != "" {
				update.SetUpdateTime(occData.UpdateTime)
			}
			if occData.VulnerabilityJSON != nil {
				update.SetVulnerabilityJSON(occData.VulnerabilityJSON)
			}
			if occData.BuildJSON != nil {
				update.SetBuildJSON(occData.BuildJSON)
			}
			if occData.ImageJSON != nil {
				update.SetImageJSON(occData.ImageJSON)
			}
			if occData.PackageJSON != nil {
				update.SetPackageJSON(occData.PackageJSON)
			}
			if occData.DeploymentJSON != nil {
				update.SetDeploymentJSON(occData.DeploymentJSON)
			}
			if occData.DiscoveryJSON != nil {
				update.SetDiscoveryJSON(occData.DiscoveryJSON)
			}
			if occData.AttestationJSON != nil {
				update.SetAttestationJSON(occData.AttestationJSON)
			}
			if occData.UpgradeJSON != nil {
				update.SetUpgradeJSON(occData.UpgradeJSON)
			}
			if occData.ComplianceJSON != nil {
				update.SetComplianceJSON(occData.ComplianceJSON)
			}
			if occData.DsseAttestationJSON != nil {
				update.SetDsseAttestationJSON(occData.DsseAttestationJSON)
			}
			if occData.SbomReferenceJSON != nil {
				update.SetSbomReferenceJSON(occData.SbomReferenceJSON)
			}
			if occData.EnvelopeJSON != nil {
				update.SetEnvelopeJSON(occData.EnvelopeJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update occurrence %s: %w", occData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, occData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for occurrence %s: %w", occData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, occData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for occurrence %s: %w", occData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleOccurrences removes occurrences that were not collected in the latest run for a project.
func (s *Service) DeleteStaleOccurrences(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	staleOccurrences, err := tx.BronzeGCPContainerAnalysisOccurrence.Query().
		Where(
			bronzegcpcontaineranalysisoccurrence.ProjectID(projectID),
			bronzegcpcontaineranalysisoccurrence.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, occ := range staleOccurrences {
		if err := s.history.CloseHistory(ctx, tx, occ.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for occurrence %s: %w", occ.ID, err)
		}

		if err := tx.BronzeGCPContainerAnalysisOccurrence.DeleteOne(occ).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete occurrence %s: %w", occ.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
