package note

import (
	"context"
	"fmt"
	"time"

	entcontaineranalysis "github.com/dannyota/hotpot/pkg/storage/ent/gcp/containeranalysis"
	"github.com/dannyota/hotpot/pkg/storage/ent/gcp/containeranalysis/bronzehistorygcpcontaineranalysisnote"
)

// HistoryService manages Grafeas note history tracking.
type HistoryService struct {
	entClient *entcontaineranalysis.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entcontaineranalysis.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new Grafeas note.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entcontaineranalysis.Tx, data *NoteData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPContainerAnalysisNote.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetShortDescription(data.ShortDescription).
		SetLongDescription(data.LongDescription).
		SetKind(data.Kind).
		SetExpirationTime(data.ExpirationTime).
		SetCreateTime(data.CreateTime).
		SetUpdateTime(data.UpdateTime).
		SetRelatedURLJSON(data.RelatedURLJSON).
		SetRelatedNoteNames(data.RelatedNoteNames).
		SetVulnerabilityJSON(data.VulnerabilityJSON).
		SetBuildJSON(data.BuildJSON).
		SetImageJSON(data.ImageJSON).
		SetPackageJSON(data.PackageJSON).
		SetDeploymentJSON(data.DeploymentJSON).
		SetDiscoveryJSON(data.DiscoveryJSON).
		SetAttestationJSON(data.AttestationJSON).
		SetUpgradeJSON(data.UpgradeJSON).
		SetComplianceJSON(data.ComplianceJSON).
		SetDsseAttestationJSON(data.DsseAttestationJSON).
		SetVulnerabilityAssessmentJSON(data.VulnerabilityAssessmentJSON).
		SetSbomReferenceJSON(data.SbomReferenceJSON).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create note history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed Grafeas note.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entcontaineranalysis.Tx, old *entcontaineranalysis.BronzeGCPContainerAnalysisNote, new *NoteData, diff *NoteDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPContainerAnalysisNote.Query().
		Where(
			bronzehistorygcpcontaineranalysisnote.ResourceID(old.ID),
			bronzehistorygcpcontaineranalysisnote.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current note history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPContainerAnalysisNote.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current note history: %w", err)
		}

		// Create new history
		_, err := tx.BronzeHistoryGCPContainerAnalysisNote.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetShortDescription(new.ShortDescription).
			SetLongDescription(new.LongDescription).
			SetKind(new.Kind).
			SetExpirationTime(new.ExpirationTime).
			SetCreateTime(new.CreateTime).
			SetUpdateTime(new.UpdateTime).
			SetRelatedURLJSON(new.RelatedURLJSON).
			SetRelatedNoteNames(new.RelatedNoteNames).
			SetVulnerabilityJSON(new.VulnerabilityJSON).
			SetBuildJSON(new.BuildJSON).
			SetImageJSON(new.ImageJSON).
			SetPackageJSON(new.PackageJSON).
			SetDeploymentJSON(new.DeploymentJSON).
			SetDiscoveryJSON(new.DiscoveryJSON).
			SetAttestationJSON(new.AttestationJSON).
			SetUpgradeJSON(new.UpgradeJSON).
			SetComplianceJSON(new.ComplianceJSON).
			SetDsseAttestationJSON(new.DsseAttestationJSON).
			SetVulnerabilityAssessmentJSON(new.VulnerabilityAssessmentJSON).
			SetSbomReferenceJSON(new.SbomReferenceJSON).
			SetProjectID(new.ProjectID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new note history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted Grafeas note.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entcontaineranalysis.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPContainerAnalysisNote.Query().
		Where(
			bronzehistorygcpcontaineranalysisnote.ResourceID(resourceID),
			bronzehistorygcpcontaineranalysisnote.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entcontaineranalysis.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current note history: %w", err)
	}

	err = tx.BronzeHistoryGCPContainerAnalysisNote.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close note history: %w", err)
	}

	return nil
}
