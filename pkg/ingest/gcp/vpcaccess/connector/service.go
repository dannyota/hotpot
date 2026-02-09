package connector

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpvpcaccessconnector"
)

// Service handles GCP VPC Access connector ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new connector ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for connector ingestion.
type IngestParams struct {
	ProjectID string
	Regions   []string
}

// IngestResult contains the result of connector ingestion.
type IngestResult struct {
	ProjectID      string
	ConnectorCount int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches connectors from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch connectors from GCP across regions
	connectors, err := s.client.ListConnectors(ctx, params.ProjectID, params.Regions)
	if err != nil {
		return nil, fmt.Errorf("failed to list connectors: %w", err)
	}

	// Convert to data structs
	connectorDataList := make([]*ConnectorData, 0, len(connectors))
	for _, c := range connectors {
		data, err := ConvertConnector(c, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert connector: %w", err)
		}
		connectorDataList = append(connectorDataList, data)
	}

	// Save to database
	if err := s.saveConnectors(ctx, connectorDataList); err != nil {
		return nil, fmt.Errorf("failed to save connectors: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		ConnectorCount: len(connectorDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveConnectors saves connectors to the database with history tracking.
func (s *Service) saveConnectors(ctx context.Context, connectors []*ConnectorData) error {
	if len(connectors) == 0 {
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

	for _, connectorData := range connectors {
		// Load existing connector
		existing, err := tx.BronzeGCPVPCAccessConnector.Query().
			Where(bronzegcpvpcaccessconnector.ID(connectorData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing connector %s: %w", connectorData.ID, err)
		}

		// Compute diff
		diff := DiffConnectorData(existing, connectorData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPVPCAccessConnector.UpdateOneID(connectorData.ID).
				SetCollectedAt(connectorData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for connector %s: %w", connectorData.ID, err)
			}
			continue
		}

		// Create or update connector
		if existing == nil {
			// Create new connector
			create := tx.BronzeGCPVPCAccessConnector.Create().
				SetID(connectorData.ID).
				SetProjectID(connectorData.ProjectID).
				SetCollectedAt(connectorData.CollectedAt).
				SetFirstCollectedAt(connectorData.CollectedAt)

			if connectorData.Network != "" {
				create.SetNetwork(connectorData.Network)
			}
			if connectorData.IpCidrRange != "" {
				create.SetIPCidrRange(connectorData.IpCidrRange)
			}
			if connectorData.State != "" {
				create.SetState(connectorData.State)
			}
			if connectorData.MinThroughput != 0 {
				create.SetMinThroughput(connectorData.MinThroughput)
			}
			if connectorData.MaxThroughput != 0 {
				create.SetMaxThroughput(connectorData.MaxThroughput)
			}
			if connectorData.MinInstances != 0 {
				create.SetMinInstances(connectorData.MinInstances)
			}
			if connectorData.MaxInstances != 0 {
				create.SetMaxInstances(connectorData.MaxInstances)
			}
			if connectorData.MachineType != "" {
				create.SetMachineType(connectorData.MachineType)
			}
			if connectorData.Region != "" {
				create.SetRegion(connectorData.Region)
			}
			if connectorData.SubnetJSON != nil {
				create.SetSubnetJSON(connectorData.SubnetJSON)
			}
			if connectorData.ConnectedProjectsJSON != nil {
				create.SetConnectedProjectsJSON(connectorData.ConnectedProjectsJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create connector %s: %w", connectorData.ID, err)
			}
		} else {
			// Update existing connector
			update := tx.BronzeGCPVPCAccessConnector.UpdateOneID(connectorData.ID).
				SetProjectID(connectorData.ProjectID).
				SetCollectedAt(connectorData.CollectedAt)

			if connectorData.Network != "" {
				update.SetNetwork(connectorData.Network)
			}
			if connectorData.IpCidrRange != "" {
				update.SetIPCidrRange(connectorData.IpCidrRange)
			}
			if connectorData.State != "" {
				update.SetState(connectorData.State)
			}
			if connectorData.MinThroughput != 0 {
				update.SetMinThroughput(connectorData.MinThroughput)
			}
			if connectorData.MaxThroughput != 0 {
				update.SetMaxThroughput(connectorData.MaxThroughput)
			}
			if connectorData.MinInstances != 0 {
				update.SetMinInstances(connectorData.MinInstances)
			}
			if connectorData.MaxInstances != 0 {
				update.SetMaxInstances(connectorData.MaxInstances)
			}
			if connectorData.MachineType != "" {
				update.SetMachineType(connectorData.MachineType)
			}
			if connectorData.Region != "" {
				update.SetRegion(connectorData.Region)
			}
			if connectorData.SubnetJSON != nil {
				update.SetSubnetJSON(connectorData.SubnetJSON)
			}
			if connectorData.ConnectedProjectsJSON != nil {
				update.SetConnectedProjectsJSON(connectorData.ConnectedProjectsJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update connector %s: %w", connectorData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, connectorData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for connector %s: %w", connectorData.ID, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, connectorData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for connector %s: %w", connectorData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleConnectors removes connectors that were not collected in the latest run.
// Also closes history records for deleted connectors.
func (s *Service) DeleteStaleConnectors(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale connectors
	staleConnectors, err := tx.BronzeGCPVPCAccessConnector.Query().
		Where(
			bronzegcpvpcaccessconnector.ProjectID(projectID),
			bronzegcpvpcaccessconnector.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale connector
	for _, connector := range staleConnectors {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, connector.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for connector %s: %w", connector.ID, err)
		}

		// Delete connector
		if err := tx.BronzeGCPVPCAccessConnector.DeleteOne(connector).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete connector %s: %w", connector.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
