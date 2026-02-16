package alertpolicy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpmonitoringalertpolicy"
)

// Service handles alert policy ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new alert policy ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for alert policy ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of alert policy ingestion.
type IngestResult struct {
	ProjectID         string
	AlertPolicyCount  int
	CollectedAt       time.Time
	DurationMillis    int64
}

// Ingest fetches alert policies from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch alert policies from GCP
	rawPolicies, err := s.client.ListAlertPolicies(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list alert policies: %w", err)
	}

	// Convert to data structs
	policyDataList := make([]*AlertPolicyData, 0, len(rawPolicies))
	for _, raw := range rawPolicies {
		data := ConvertAlertPolicy(raw, params.ProjectID, collectedAt)
		policyDataList = append(policyDataList, data)
	}

	// Save to database
	if err := s.saveAlertPolicies(ctx, policyDataList); err != nil {
		return nil, fmt.Errorf("failed to save alert policies: %w", err)
	}

	return &IngestResult{
		ProjectID:        params.ProjectID,
		AlertPolicyCount: len(policyDataList),
		CollectedAt:      collectedAt,
		DurationMillis:   time.Since(startTime).Milliseconds(),
	}, nil
}

// saveAlertPolicies saves alert policies to the database with history tracking.
func (s *Service) saveAlertPolicies(ctx context.Context, policies []*AlertPolicyData) error {
	if len(policies) == 0 {
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

	for _, policyData := range policies {
		// Load existing alert policy
		existing, err := tx.BronzeGCPMonitoringAlertPolicy.Query().
			Where(bronzegcpmonitoringalertpolicy.ID(policyData.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing alert policy %s: %w", policyData.ID, err)
		}

		// Compute diff
		diff := DiffAlertPolicyData(existing, policyData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPMonitoringAlertPolicy.UpdateOneID(policyData.ID).
				SetCollectedAt(policyData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for alert policy %s: %w", policyData.ID, err)
			}
			continue
		}

		// Create or update alert policy
		if existing == nil {
			create := tx.BronzeGCPMonitoringAlertPolicy.Create().
				SetID(policyData.ID).
				SetName(policyData.Name).
				SetCombiner(policyData.Combiner).
				SetEnabled(policyData.Enabled).
				SetSeverity(policyData.Severity).
				SetProjectID(policyData.ProjectID).
				SetCollectedAt(policyData.CollectedAt).
				SetFirstCollectedAt(policyData.CollectedAt)

			if policyData.DisplayName != "" {
				create.SetDisplayName(policyData.DisplayName)
			}
			if policyData.DocumentationJSON != nil {
				create.SetDocumentationJSON(policyData.DocumentationJSON)
			}
			if policyData.UserLabelsJSON != nil {
				create.SetUserLabelsJSON(policyData.UserLabelsJSON)
			}
			if policyData.ConditionsJSON != nil {
				create.SetConditionsJSON(policyData.ConditionsJSON)
			}
			if policyData.NotificationChannelsJSON != nil {
				create.SetNotificationChannelsJSON(policyData.NotificationChannelsJSON)
			}
			if policyData.CreationRecordJSON != nil {
				create.SetCreationRecordJSON(policyData.CreationRecordJSON)
			}
			if policyData.MutationRecordJSON != nil {
				create.SetMutationRecordJSON(policyData.MutationRecordJSON)
			}
			if policyData.AlertStrategyJSON != nil {
				create.SetAlertStrategyJSON(policyData.AlertStrategyJSON)
			}

			_, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create alert policy %s: %w", policyData.ID, err)
			}
		} else {
			update := tx.BronzeGCPMonitoringAlertPolicy.UpdateOneID(policyData.ID).
				SetName(policyData.Name).
				SetCombiner(policyData.Combiner).
				SetEnabled(policyData.Enabled).
				SetSeverity(policyData.Severity).
				SetProjectID(policyData.ProjectID).
				SetCollectedAt(policyData.CollectedAt)

			if policyData.DisplayName != "" {
				update.SetDisplayName(policyData.DisplayName)
			}
			if policyData.DocumentationJSON != nil {
				update.SetDocumentationJSON(policyData.DocumentationJSON)
			}
			if policyData.UserLabelsJSON != nil {
				update.SetUserLabelsJSON(policyData.UserLabelsJSON)
			}
			if policyData.ConditionsJSON != nil {
				update.SetConditionsJSON(policyData.ConditionsJSON)
			}
			if policyData.NotificationChannelsJSON != nil {
				update.SetNotificationChannelsJSON(policyData.NotificationChannelsJSON)
			}
			if policyData.CreationRecordJSON != nil {
				update.SetCreationRecordJSON(policyData.CreationRecordJSON)
			}
			if policyData.MutationRecordJSON != nil {
				update.SetMutationRecordJSON(policyData.MutationRecordJSON)
			}
			if policyData.AlertStrategyJSON != nil {
				update.SetAlertStrategyJSON(policyData.AlertStrategyJSON)
			}

			_, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update alert policy %s: %w", policyData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, policyData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for alert policy %s: %w", policyData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, policyData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for alert policy %s: %w", policyData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleAlertPolicies removes alert policies that were not collected in the latest run.
func (s *Service) DeleteStaleAlertPolicies(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	stalePolicies, err := tx.BronzeGCPMonitoringAlertPolicy.Query().
		Where(
			bronzegcpmonitoringalertpolicy.ProjectID(projectID),
			bronzegcpmonitoringalertpolicy.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, policy := range stalePolicies {
		if err := s.history.CloseHistory(ctx, tx, policy.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for alert policy %s: %w", policy.ID, err)
		}

		if err := tx.BronzeGCPMonitoringAlertPolicy.DeleteOne(policy).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete alert policy %s: %w", policy.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
