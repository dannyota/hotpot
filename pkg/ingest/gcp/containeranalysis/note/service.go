package note

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcontaineranalysisnote"
)

// Service handles Grafeas note ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Grafeas note ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of note ingestion.
type IngestResult struct {
	ProjectID      string
	NoteCount      int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches Grafeas notes from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	rawNotes, err := s.client.ListNotes(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list notes: %w", err)
	}

	noteDataList := make([]*NoteData, 0, len(rawNotes))
	for _, raw := range rawNotes {
		data := ConvertNote(raw, projectID, collectedAt)
		noteDataList = append(noteDataList, data)
	}

	if err := s.saveNotes(ctx, noteDataList); err != nil {
		return nil, fmt.Errorf("failed to save notes: %w", err)
	}

	return &IngestResult{
		ProjectID:      projectID,
		NoteCount:      len(noteDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveNotes saves Grafeas notes to the database with history tracking.
func (s *Service) saveNotes(ctx context.Context, notes []*NoteData) error {
	if len(notes) == 0 {
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

	for _, noteData := range notes {
		// Load existing note
		existing, err := tx.BronzeGCPContainerAnalysisNote.Query().
			Where(bronzegcpcontaineranalysisnote.ID(noteData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing note %s: %w", noteData.ID, err)
		}

		// Compute diff
		diff := DiffNoteData(existing, noteData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPContainerAnalysisNote.UpdateOneID(noteData.ID).
				SetCollectedAt(noteData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for note %s: %w", noteData.ID, err)
			}
			continue
		}

		// Create or update note
		if existing == nil {
			create := tx.BronzeGCPContainerAnalysisNote.Create().
				SetID(noteData.ID).
				SetProjectID(noteData.ProjectID).
				SetCollectedAt(noteData.CollectedAt).
				SetFirstCollectedAt(noteData.CollectedAt).
				SetKind(noteData.Kind)

			if noteData.ShortDescription != "" {
				create.SetShortDescription(noteData.ShortDescription)
			}
			if noteData.LongDescription != "" {
				create.SetLongDescription(noteData.LongDescription)
			}
			if noteData.ExpirationTime != "" {
				create.SetExpirationTime(noteData.ExpirationTime)
			}
			if noteData.CreateTime != "" {
				create.SetCreateTime(noteData.CreateTime)
			}
			if noteData.UpdateTime != "" {
				create.SetUpdateTime(noteData.UpdateTime)
			}
			if noteData.RelatedURLJSON != nil {
				create.SetRelatedURLJSON(noteData.RelatedURLJSON)
			}
			if noteData.RelatedNoteNames != nil {
				create.SetRelatedNoteNames(noteData.RelatedNoteNames)
			}
			if noteData.VulnerabilityJSON != nil {
				create.SetVulnerabilityJSON(noteData.VulnerabilityJSON)
			}
			if noteData.BuildJSON != nil {
				create.SetBuildJSON(noteData.BuildJSON)
			}
			if noteData.ImageJSON != nil {
				create.SetImageJSON(noteData.ImageJSON)
			}
			if noteData.PackageJSON != nil {
				create.SetPackageJSON(noteData.PackageJSON)
			}
			if noteData.DeploymentJSON != nil {
				create.SetDeploymentJSON(noteData.DeploymentJSON)
			}
			if noteData.DiscoveryJSON != nil {
				create.SetDiscoveryJSON(noteData.DiscoveryJSON)
			}
			if noteData.AttestationJSON != nil {
				create.SetAttestationJSON(noteData.AttestationJSON)
			}
			if noteData.UpgradeJSON != nil {
				create.SetUpgradeJSON(noteData.UpgradeJSON)
			}
			if noteData.ComplianceJSON != nil {
				create.SetComplianceJSON(noteData.ComplianceJSON)
			}
			if noteData.DsseAttestationJSON != nil {
				create.SetDsseAttestationJSON(noteData.DsseAttestationJSON)
			}
			if noteData.VulnerabilityAssessmentJSON != nil {
				create.SetVulnerabilityAssessmentJSON(noteData.VulnerabilityAssessmentJSON)
			}
			if noteData.SbomReferenceJSON != nil {
				create.SetSbomReferenceJSON(noteData.SbomReferenceJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create note %s: %w", noteData.ID, err)
			}
		} else {
			update := tx.BronzeGCPContainerAnalysisNote.UpdateOneID(noteData.ID).
				SetProjectID(noteData.ProjectID).
				SetCollectedAt(noteData.CollectedAt).
				SetKind(noteData.Kind)

			if noteData.ShortDescription != "" {
				update.SetShortDescription(noteData.ShortDescription)
			}
			if noteData.LongDescription != "" {
				update.SetLongDescription(noteData.LongDescription)
			}
			if noteData.ExpirationTime != "" {
				update.SetExpirationTime(noteData.ExpirationTime)
			}
			if noteData.CreateTime != "" {
				update.SetCreateTime(noteData.CreateTime)
			}
			if noteData.UpdateTime != "" {
				update.SetUpdateTime(noteData.UpdateTime)
			}
			if noteData.RelatedURLJSON != nil {
				update.SetRelatedURLJSON(noteData.RelatedURLJSON)
			}
			if noteData.RelatedNoteNames != nil {
				update.SetRelatedNoteNames(noteData.RelatedNoteNames)
			}
			if noteData.VulnerabilityJSON != nil {
				update.SetVulnerabilityJSON(noteData.VulnerabilityJSON)
			}
			if noteData.BuildJSON != nil {
				update.SetBuildJSON(noteData.BuildJSON)
			}
			if noteData.ImageJSON != nil {
				update.SetImageJSON(noteData.ImageJSON)
			}
			if noteData.PackageJSON != nil {
				update.SetPackageJSON(noteData.PackageJSON)
			}
			if noteData.DeploymentJSON != nil {
				update.SetDeploymentJSON(noteData.DeploymentJSON)
			}
			if noteData.DiscoveryJSON != nil {
				update.SetDiscoveryJSON(noteData.DiscoveryJSON)
			}
			if noteData.AttestationJSON != nil {
				update.SetAttestationJSON(noteData.AttestationJSON)
			}
			if noteData.UpgradeJSON != nil {
				update.SetUpgradeJSON(noteData.UpgradeJSON)
			}
			if noteData.ComplianceJSON != nil {
				update.SetComplianceJSON(noteData.ComplianceJSON)
			}
			if noteData.DsseAttestationJSON != nil {
				update.SetDsseAttestationJSON(noteData.DsseAttestationJSON)
			}
			if noteData.VulnerabilityAssessmentJSON != nil {
				update.SetVulnerabilityAssessmentJSON(noteData.VulnerabilityAssessmentJSON)
			}
			if noteData.SbomReferenceJSON != nil {
				update.SetSbomReferenceJSON(noteData.SbomReferenceJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update note %s: %w", noteData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, noteData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for note %s: %w", noteData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, noteData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for note %s: %w", noteData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleNotes removes notes that were not collected in the latest run for a project.
func (s *Service) DeleteStaleNotes(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleNotes, err := tx.BronzeGCPContainerAnalysisNote.Query().
		Where(
			bronzegcpcontaineranalysisnote.ProjectID(projectID),
			bronzegcpcontaineranalysisnote.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, note := range staleNotes {
		if err := s.history.CloseHistory(ctx, tx, note.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for note %s: %w", note.ID, err)
		}

		if err := tx.BronzeGCPContainerAnalysisNote.DeleteOne(note).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete note %s: %w", note.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
