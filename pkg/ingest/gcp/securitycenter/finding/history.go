package finding

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpsecuritycenterfinding"
)

// HistoryService manages SCC finding history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new SCC finding.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *FindingData, now time.Time) error {
	create := tx.BronzeHistoryGCPSecurityCenterFinding.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetParent(data.Parent).
		SetOrganizationID(data.OrganizationID)

	if data.ResourceName != "" {
		create.SetResourceName(data.ResourceName)
	}
	if data.State != "" {
		create.SetState(data.State)
	}
	if data.Category != "" {
		create.SetCategory(data.Category)
	}
	if data.ExternalURI != "" {
		create.SetExternalURI(data.ExternalURI)
	}
	if data.Severity != "" {
		create.SetSeverity(data.Severity)
	}
	if data.FindingClass != "" {
		create.SetFindingClass(data.FindingClass)
	}
	if data.CanonicalName != "" {
		create.SetCanonicalName(data.CanonicalName)
	}
	if data.Mute != "" {
		create.SetMute(data.Mute)
	}
	if data.CreateTime != "" {
		create.SetCreateTime(data.CreateTime)
	}
	if data.EventTime != "" {
		create.SetEventTime(data.EventTime)
	}
	if data.SourceProperties != nil {
		create.SetSourceProperties(data.SourceProperties)
	}
	if data.SecurityMarks != nil {
		create.SetSecurityMarks(data.SecurityMarks)
	}
	if data.Indicator != nil {
		create.SetIndicator(data.Indicator)
	}
	if data.Vulnerability != nil {
		create.SetVulnerability(data.Vulnerability)
	}
	if data.Connections != nil {
		create.SetConnections(data.Connections)
	}
	if data.Compliances != nil {
		create.SetCompliances(data.Compliances)
	}
	if data.Contacts != nil {
		create.SetContacts(data.Contacts)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("failed to create SCC finding history: %w", err)
	}

	return nil
}

// UpdateHistory updates history records for a changed SCC finding.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPSecurityCenterFinding, new *FindingData, diff *FindingDiff, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPSecurityCenterFinding.Query().
		Where(
			bronzehistorygcpsecuritycenterfinding.ResourceID(old.ID),
			bronzehistorygcpsecuritycenterfinding.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current SCC finding history: %w", err)
	}

	if diff.IsChanged {
		// Close current history
		err = tx.BronzeHistoryGCPSecurityCenterFinding.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current SCC finding history: %w", err)
		}

		// Create new history
		create := tx.BronzeHistoryGCPSecurityCenterFinding.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetParent(new.Parent).
			SetOrganizationID(new.OrganizationID)

		if new.ResourceName != "" {
			create.SetResourceName(new.ResourceName)
		}
		if new.State != "" {
			create.SetState(new.State)
		}
		if new.Category != "" {
			create.SetCategory(new.Category)
		}
		if new.ExternalURI != "" {
			create.SetExternalURI(new.ExternalURI)
		}
		if new.Severity != "" {
			create.SetSeverity(new.Severity)
		}
		if new.FindingClass != "" {
			create.SetFindingClass(new.FindingClass)
		}
		if new.CanonicalName != "" {
			create.SetCanonicalName(new.CanonicalName)
		}
		if new.Mute != "" {
			create.SetMute(new.Mute)
		}
		if new.CreateTime != "" {
			create.SetCreateTime(new.CreateTime)
		}
		if new.EventTime != "" {
			create.SetEventTime(new.EventTime)
		}
		if new.SourceProperties != nil {
			create.SetSourceProperties(new.SourceProperties)
		}
		if new.SecurityMarks != nil {
			create.SetSecurityMarks(new.SecurityMarks)
		}
		if new.Indicator != nil {
			create.SetIndicator(new.Indicator)
		}
		if new.Vulnerability != nil {
			create.SetVulnerability(new.Vulnerability)
		}
		if new.Connections != nil {
			create.SetConnections(new.Connections)
		}
		if new.Compliances != nil {
			create.SetCompliances(new.Compliances)
		}
		if new.Contacts != nil {
			create.SetContacts(new.Contacts)
		}

		if _, err := create.Save(ctx); err != nil {
			return fmt.Errorf("failed to create new SCC finding history: %w", err)
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted SCC finding.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHistory, err := tx.BronzeHistoryGCPSecurityCenterFinding.Query().
		Where(
			bronzehistorygcpsecuritycenterfinding.ResourceID(resourceID),
			bronzehistorygcpsecuritycenterfinding.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to find current SCC finding history: %w", err)
	}

	err = tx.BronzeHistoryGCPSecurityCenterFinding.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close SCC finding history: %w", err)
	}

	return nil
}
