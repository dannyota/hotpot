package interconnect

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputeinterconnect"
)

// Service handles GCP Compute interconnect ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new interconnect ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for interconnect ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of interconnect ingestion.
type IngestResult struct {
	ProjectID          string
	InterconnectCount  int
	CollectedAt        time.Time
	DurationMillis     int64
}

// Ingest fetches interconnects from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch interconnects from GCP
	interconnects, err := s.client.ListInterconnects(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list interconnects: %w", err)
	}

	// Convert to data structs
	interconnectDataList := make([]*InterconnectData, 0, len(interconnects))
	for _, ic := range interconnects {
		data, err := ConvertInterconnect(ic, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert interconnect: %w", err)
		}
		interconnectDataList = append(interconnectDataList, data)
	}

	// Save to database
	if err := s.saveInterconnects(ctx, interconnectDataList); err != nil {
		return nil, fmt.Errorf("failed to save interconnects: %w", err)
	}

	return &IngestResult{
		ProjectID:         params.ProjectID,
		InterconnectCount: len(interconnectDataList),
		CollectedAt:       collectedAt,
		DurationMillis:    time.Since(startTime).Milliseconds(),
	}, nil
}

// saveInterconnects saves interconnects to the database with history tracking.
func (s *Service) saveInterconnects(ctx context.Context, interconnects []*InterconnectData) error {
	if len(interconnects) == 0 {
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

	for _, icData := range interconnects {
		// Load existing interconnect
		existing, err := tx.BronzeGCPComputeInterconnect.Query().
			Where(bronzegcpcomputeinterconnect.ID(icData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing interconnect %s: %w", icData.Name, err)
		}

		// Compute diff
		diff := DiffInterconnectData(existing, icData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeInterconnect.UpdateOneID(icData.ID).
				SetCollectedAt(icData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for interconnect %s: %w", icData.Name, err)
			}
			continue
		}

		// Create or update interconnect
		if existing == nil {
			// Create new interconnect
			create := tx.BronzeGCPComputeInterconnect.Create().
				SetID(icData.ID).
				SetName(icData.Name).
				SetProjectID(icData.ProjectID).
				SetCollectedAt(icData.CollectedAt).
				SetFirstCollectedAt(icData.CollectedAt)

			if icData.Description != "" {
				create.SetDescription(icData.Description)
			}
			if icData.SelfLink != "" {
				create.SetSelfLink(icData.SelfLink)
			}
			if icData.Location != "" {
				create.SetLocation(icData.Location)
			}
			if icData.InterconnectType != "" {
				create.SetInterconnectType(icData.InterconnectType)
			}
			if icData.LinkType != "" {
				create.SetLinkType(icData.LinkType)
			}
			if icData.AdminEnabled {
				create.SetAdminEnabled(icData.AdminEnabled)
			}
			if icData.OperationalStatus != "" {
				create.SetOperationalStatus(icData.OperationalStatus)
			}
			if icData.ProvisionedLinkCount != 0 {
				create.SetProvisionedLinkCount(icData.ProvisionedLinkCount)
			}
			if icData.RequestedLinkCount != 0 {
				create.SetRequestedLinkCount(icData.RequestedLinkCount)
			}
			if icData.PeerIPAddress != "" {
				create.SetPeerIPAddress(icData.PeerIPAddress)
			}
			if icData.GoogleIPAddress != "" {
				create.SetGoogleIPAddress(icData.GoogleIPAddress)
			}
			if icData.GoogleReferenceID != "" {
				create.SetGoogleReferenceID(icData.GoogleReferenceID)
			}
			if icData.NocContactEmail != "" {
				create.SetNocContactEmail(icData.NocContactEmail)
			}
			if icData.CustomerName != "" {
				create.SetCustomerName(icData.CustomerName)
			}
			if icData.State != "" {
				create.SetState(icData.State)
			}
			if icData.CreationTimestamp != "" {
				create.SetCreationTimestamp(icData.CreationTimestamp)
			}
			if icData.ExpectedOutagesJSON != nil {
				create.SetExpectedOutagesJSON(icData.ExpectedOutagesJSON)
			}
			if icData.CircuitInfosJSON != nil {
				create.SetCircuitInfosJSON(icData.CircuitInfosJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create interconnect %s: %w", icData.Name, err)
			}
		} else {
			// Update existing interconnect
			update := tx.BronzeGCPComputeInterconnect.UpdateOneID(icData.ID).
				SetName(icData.Name).
				SetProjectID(icData.ProjectID).
				SetCollectedAt(icData.CollectedAt)

			if icData.Description != "" {
				update.SetDescription(icData.Description)
			}
			if icData.SelfLink != "" {
				update.SetSelfLink(icData.SelfLink)
			}
			if icData.Location != "" {
				update.SetLocation(icData.Location)
			}
			if icData.InterconnectType != "" {
				update.SetInterconnectType(icData.InterconnectType)
			}
			if icData.LinkType != "" {
				update.SetLinkType(icData.LinkType)
			}
			if icData.AdminEnabled {
				update.SetAdminEnabled(icData.AdminEnabled)
			}
			if icData.OperationalStatus != "" {
				update.SetOperationalStatus(icData.OperationalStatus)
			}
			if icData.ProvisionedLinkCount != 0 {
				update.SetProvisionedLinkCount(icData.ProvisionedLinkCount)
			}
			if icData.RequestedLinkCount != 0 {
				update.SetRequestedLinkCount(icData.RequestedLinkCount)
			}
			if icData.PeerIPAddress != "" {
				update.SetPeerIPAddress(icData.PeerIPAddress)
			}
			if icData.GoogleIPAddress != "" {
				update.SetGoogleIPAddress(icData.GoogleIPAddress)
			}
			if icData.GoogleReferenceID != "" {
				update.SetGoogleReferenceID(icData.GoogleReferenceID)
			}
			if icData.NocContactEmail != "" {
				update.SetNocContactEmail(icData.NocContactEmail)
			}
			if icData.CustomerName != "" {
				update.SetCustomerName(icData.CustomerName)
			}
			if icData.State != "" {
				update.SetState(icData.State)
			}
			if icData.CreationTimestamp != "" {
				update.SetCreationTimestamp(icData.CreationTimestamp)
			}
			if icData.ExpectedOutagesJSON != nil {
				update.SetExpectedOutagesJSON(icData.ExpectedOutagesJSON)
			}
			if icData.CircuitInfosJSON != nil {
				update.SetCircuitInfosJSON(icData.CircuitInfosJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update interconnect %s: %w", icData.Name, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, icData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for interconnect %s: %w", icData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, icData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for interconnect %s: %w", icData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleInterconnects removes interconnects that were not collected in the latest run.
// Also closes history records for deleted interconnects.
func (s *Service) DeleteStaleInterconnects(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale interconnects
	staleInterconnects, err := tx.BronzeGCPComputeInterconnect.Query().
		Where(
			bronzegcpcomputeinterconnect.ProjectID(projectID),
			bronzegcpcomputeinterconnect.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale interconnect
	for _, ic := range staleInterconnects {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, ic.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for interconnect %s: %w", ic.ID, err)
		}

		// Delete interconnect
		if err := tx.BronzeGCPComputeInterconnect.DeleteOne(ic).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete interconnect %s: %w", ic.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
