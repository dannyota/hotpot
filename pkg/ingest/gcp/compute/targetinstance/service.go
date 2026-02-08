package targetinstance

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpcomputetargetinstance"
)

// Service handles GCP Compute target instance ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new target instance ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for target instance ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of target instance ingestion.
type IngestResult struct {
	ProjectID           string
	TargetInstanceCount int
	CollectedAt         time.Time
	DurationMillis      int64
}

// Ingest fetches target instances from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch target instances from GCP
	targetInstances, err := s.client.ListTargetInstances(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list target instances: %w", err)
	}

	// Convert to data structs
	targetInstanceDataList := make([]*TargetInstanceData, 0, len(targetInstances))
	for _, ti := range targetInstances {
		data := ConvertTargetInstance(ti, params.ProjectID, collectedAt)
		targetInstanceDataList = append(targetInstanceDataList, data)
	}

	// Save to database
	if err := s.saveTargetInstances(ctx, targetInstanceDataList); err != nil {
		return nil, fmt.Errorf("failed to save target instances: %w", err)
	}

	return &IngestResult{
		ProjectID:           params.ProjectID,
		TargetInstanceCount: len(targetInstanceDataList),
		CollectedAt:         collectedAt,
		DurationMillis:      time.Since(startTime).Milliseconds(),
	}, nil
}

// saveTargetInstances saves target instances to the database with history tracking.
func (s *Service) saveTargetInstances(ctx context.Context, instances []*TargetInstanceData) error {
	if len(instances) == 0 {
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

	for _, instanceData := range instances {
		// Load existing target instance
		existing, err := tx.BronzeGCPComputeTargetInstance.Query().
			Where(bronzegcpcomputetargetinstance.ID(instanceData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing target instance %s: %w", instanceData.Name, err)
		}

		// Compute diff
		diff := DiffTargetInstanceData(existing, instanceData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeTargetInstance.UpdateOneID(instanceData.ID).
				SetCollectedAt(instanceData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for target instance %s: %w", instanceData.Name, err)
			}
			continue
		}

		// Create or update target instance
		if existing == nil {
			// Create new target instance
			create := tx.BronzeGCPComputeTargetInstance.Create().
				SetID(instanceData.ID).
				SetName(instanceData.Name).
				SetProjectID(instanceData.ProjectID).
				SetCollectedAt(instanceData.CollectedAt)

			if instanceData.Description != "" {
				create.SetDescription(instanceData.Description)
			}
			if instanceData.Zone != "" {
				create.SetZone(instanceData.Zone)
			}
			if instanceData.Instance != "" {
				create.SetInstance(instanceData.Instance)
			}
			if instanceData.Network != "" {
				create.SetNetwork(instanceData.Network)
			}
			if instanceData.NatPolicy != "" {
				create.SetNatPolicy(instanceData.NatPolicy)
			}
			if instanceData.SecurityPolicy != "" {
				create.SetSecurityPolicy(instanceData.SecurityPolicy)
			}
			if instanceData.SelfLink != "" {
				create.SetSelfLink(instanceData.SelfLink)
			}
			if instanceData.CreationTimestamp != "" {
				create.SetCreationTimestamp(instanceData.CreationTimestamp)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create target instance %s: %w", instanceData.Name, err)
			}
		} else {
			// Update existing target instance
			update := tx.BronzeGCPComputeTargetInstance.UpdateOneID(instanceData.ID).
				SetName(instanceData.Name).
				SetProjectID(instanceData.ProjectID).
				SetCollectedAt(instanceData.CollectedAt)

			if instanceData.Description != "" {
				update.SetDescription(instanceData.Description)
			}
			if instanceData.Zone != "" {
				update.SetZone(instanceData.Zone)
			}
			if instanceData.Instance != "" {
				update.SetInstance(instanceData.Instance)
			}
			if instanceData.Network != "" {
				update.SetNetwork(instanceData.Network)
			}
			if instanceData.NatPolicy != "" {
				update.SetNatPolicy(instanceData.NatPolicy)
			}
			if instanceData.SecurityPolicy != "" {
				update.SetSecurityPolicy(instanceData.SecurityPolicy)
			}
			if instanceData.SelfLink != "" {
				update.SetSelfLink(instanceData.SelfLink)
			}
			if instanceData.CreationTimestamp != "" {
				update.SetCreationTimestamp(instanceData.CreationTimestamp)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update target instance %s: %w", instanceData.Name, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, instanceData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for target instance %s: %w", instanceData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, instanceData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for target instance %s: %w", instanceData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleTargetInstances removes target instances that were not collected in the latest run.
// Also closes history records for deleted target instances.
func (s *Service) DeleteStaleTargetInstances(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale target instances
	staleInstances, err := tx.BronzeGCPComputeTargetInstance.Query().
		Where(
			bronzegcpcomputetargetinstance.ProjectID(projectID),
			bronzegcpcomputetargetinstance.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale target instance
	for _, instance := range staleInstances {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, instance.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for target instance %s: %w", instance.ID, err)
		}

		// Delete target instance
		if err := tx.BronzeGCPComputeTargetInstance.DeleteOne(instance).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete target instance %s: %w", instance.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
