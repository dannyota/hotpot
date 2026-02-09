package threat

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorys1threat"
)

// HistoryService handles history tracking for threats.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new threat.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *ThreatData, now time.Time) error {
	create := tx.BronzeHistoryS1Threat.Create().
		SetResourceID(data.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetAgentID(data.AgentID).
		SetClassification(data.Classification).
		SetThreatName(data.ThreatName).
		SetFilePath(data.FilePath).
		SetStatus(data.Status).
		SetAnalystVerdict(data.AnalystVerdict).
		SetConfidenceLevel(data.ConfidenceLevel).
		SetInitiatedBy(data.InitiatedBy).
		SetFileContentHash(data.FileContentHash).
		SetFileSha256(data.FileSHA256).
		SetCloudVerdict(data.CloudVerdict).
		SetClassificationSource(data.ClassificationSource).
		SetSiteID(data.SiteID).
		SetSiteName(data.SiteName).
		SetAccountID(data.AccountID).
		SetAccountName(data.AccountName).
		SetAgentComputerName(data.AgentComputerName).
		SetAgentOsType(data.AgentOsType).
		SetAgentMachineType(data.AgentMachineType).
		SetAgentIsActive(data.AgentIsActive).
		SetAgentIsDecommissioned(data.AgentIsDecommissioned).
		SetAgentVersion(data.AgentVersion)

	if data.APICreatedAt != nil {
		create.SetAPICreatedAt(*data.APICreatedAt)
	}
	if data.ThreatInfoJSON != nil {
		create.SetThreatInfoJSON(data.ThreatInfoJSON)
	}
	if data.APIUpdatedAt != nil {
		create.SetAPIUpdatedAt(*data.APIUpdatedAt)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("create threat history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new history for a changed threat.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeS1Threat, new *ThreatData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1Threat.Query().
		Where(
			bronzehistorys1threat.ResourceID(old.ID),
			bronzehistorys1threat.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current threat history: %w", err)
	}

	if err := tx.BronzeHistoryS1Threat.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close threat history: %w", err)
	}

	create := tx.BronzeHistoryS1Threat.Create().
		SetResourceID(new.ResourceID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetAgentID(new.AgentID).
		SetClassification(new.Classification).
		SetThreatName(new.ThreatName).
		SetFilePath(new.FilePath).
		SetStatus(new.Status).
		SetAnalystVerdict(new.AnalystVerdict).
		SetConfidenceLevel(new.ConfidenceLevel).
		SetInitiatedBy(new.InitiatedBy).
		SetFileContentHash(new.FileContentHash).
		SetFileSha256(new.FileSHA256).
		SetCloudVerdict(new.CloudVerdict).
		SetClassificationSource(new.ClassificationSource).
		SetSiteID(new.SiteID).
		SetSiteName(new.SiteName).
		SetAccountID(new.AccountID).
		SetAccountName(new.AccountName).
		SetAgentComputerName(new.AgentComputerName).
		SetAgentOsType(new.AgentOsType).
		SetAgentMachineType(new.AgentMachineType).
		SetAgentIsActive(new.AgentIsActive).
		SetAgentIsDecommissioned(new.AgentIsDecommissioned).
		SetAgentVersion(new.AgentVersion)

	if new.APICreatedAt != nil {
		create.SetAPICreatedAt(*new.APICreatedAt)
	}
	if new.ThreatInfoJSON != nil {
		create.SetThreatInfoJSON(new.ThreatInfoJSON)
	}
	if new.APIUpdatedAt != nil {
		create.SetAPIUpdatedAt(*new.APIUpdatedAt)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("create new threat history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted threat.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryS1Threat.Query().
		Where(
			bronzehistorys1threat.ResourceID(resourceID),
			bronzehistorys1threat.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current threat history: %w", err)
	}

	if err := tx.BronzeHistoryS1Threat.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close threat history: %w", err)
	}

	return nil
}
