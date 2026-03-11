package occurrence

import (
	"context"
	"fmt"
	"time"

	entcontaineranalysis "danny.vn/hotpot/pkg/storage/ent/gcp/containeranalysis"
	"danny.vn/hotpot/pkg/storage/ent/gcp/containeranalysis/bronzehistorygcpcontaineranalysisoccurrence"
)

// HistoryService manages Grafeas occurrence history tracking.
type HistoryService struct {
	entClient *entcontaineranalysis.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *entcontaineranalysis.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new Grafeas occurrence.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *entcontaineranalysis.Tx, data *OccurrenceData, now time.Time) error {
	_, err := tx.BronzeHistoryGCPContainerAnalysisOccurrence.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetResourceURI(data.ResourceURI).
		SetNoteName(data.NoteName).
		SetKind(data.Kind).
		SetRemediation(data.Remediation).
		SetCreateTime(data.CreateTime).
		SetUpdateTime(data.UpdateTime).
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
		SetSbomReferenceJSON(data.SbomReferenceJSON).
		SetEnvelopeJSON(data.EnvelopeJSON).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create occurrence history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed Grafeas occurrence.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *entcontaineranalysis.Tx, old *entcontaineranalysis.BronzeGCPContainerAnalysisOccurrence, new *OccurrenceData, diff *OccurrenceDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPContainerAnalysisOccurrence.Query().
		Where(
			bronzehistorygcpcontaineranalysisoccurrence.ResourceID(old.ID),
			bronzehistorygcpcontaineranalysisoccurrence.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current occurrence history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPContainerAnalysisOccurrence.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current occurrence history: %w", err)
		}

		// Create new history
		_, err := tx.BronzeHistoryGCPContainerAnalysisOccurrence.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetResourceURI(new.ResourceURI).
			SetNoteName(new.NoteName).
			SetKind(new.Kind).
			SetRemediation(new.Remediation).
			SetCreateTime(new.CreateTime).
			SetUpdateTime(new.UpdateTime).
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
			SetSbomReferenceJSON(new.SbomReferenceJSON).
			SetEnvelopeJSON(new.EnvelopeJSON).
			SetProjectID(new.ProjectID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new occurrence history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted Grafeas occurrence.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *entcontaineranalysis.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPContainerAnalysisOccurrence.Query().
		Where(
			bronzehistorygcpcontaineranalysisoccurrence.ResourceID(resourceID),
			bronzehistorygcpcontaineranalysisoccurrence.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if entcontaineranalysis.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current occurrence history: %w", err)
	}

	err = tx.BronzeHistoryGCPContainerAnalysisOccurrence.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close occurrence history: %w", err)
	}

	return nil
}
