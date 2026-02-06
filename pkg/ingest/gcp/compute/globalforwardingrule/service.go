package globalforwardingrule

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute global forwarding rule ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new global forwarding rule ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for global forwarding rule ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of global forwarding rule ingestion.
type IngestResult struct {
	ProjectID                string
	GlobalForwardingRuleCount int
	CollectedAt              time.Time
	DurationMillis           int64
}

// Ingest fetches global forwarding rules from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch global forwarding rules from GCP
	forwardingRules, err := s.client.ListGlobalForwardingRules(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list global forwarding rules: %w", err)
	}

	// Convert to bronze models
	bronzeRules := make([]bronze.GCPComputeGlobalForwardingRule, 0, len(forwardingRules))
	for _, fr := range forwardingRules {
		rule, err := ConvertGlobalForwardingRule(fr, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert global forwarding rule: %w", err)
		}
		bronzeRules = append(bronzeRules, rule)
	}

	// Save to database
	if err := s.saveGlobalForwardingRules(ctx, bronzeRules); err != nil {
		return nil, fmt.Errorf("failed to save global forwarding rules: %w", err)
	}

	return &IngestResult{
		ProjectID:                params.ProjectID,
		GlobalForwardingRuleCount: len(bronzeRules),
		CollectedAt:              collectedAt,
		DurationMillis:           time.Since(startTime).Milliseconds(),
	}, nil
}

// saveGlobalForwardingRules saves global forwarding rules to the database with history tracking.
func (s *Service) saveGlobalForwardingRules(ctx context.Context, forwardingRules []bronze.GCPComputeGlobalForwardingRule) error {
	if len(forwardingRules) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, rule := range forwardingRules {
			// Load existing global forwarding rule with all relations
			var existing *bronze.GCPComputeGlobalForwardingRule
			var old bronze.GCPComputeGlobalForwardingRule
			err := tx.Preload("Labels").
				Where("resource_id = ?", rule.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing global forwarding rule %s: %w", rule.Name, err)
			}

			// Compute diff
			diff := DiffGlobalForwardingRule(existing, &rule)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeGlobalForwardingRule{}).
					Where("resource_id = ?", rule.ResourceID).
					Update("collected_at", rule.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for global forwarding rule %s: %w", rule.Name, err)
				}
				continue
			}

			// Delete old relations (manual cascade)
			if existing != nil {
				if err := s.deleteGlobalForwardingRuleRelations(tx, rule.ResourceID); err != nil {
					return fmt.Errorf("failed to delete old relations for global forwarding rule %s: %w", rule.Name, err)
				}
			}

			// Upsert global forwarding rule
			if err := tx.Save(&rule).Error; err != nil {
				return fmt.Errorf("failed to upsert global forwarding rule %s: %w", rule.Name, err)
			}

			// Create new relations
			if err := s.createGlobalForwardingRuleRelations(tx, rule.ResourceID, &rule); err != nil {
				return fmt.Errorf("failed to create relations for global forwarding rule %s: %w", rule.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &rule, now); err != nil {
					return fmt.Errorf("failed to create history for global forwarding rule %s: %w", rule.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &rule, diff, now); err != nil {
					return fmt.Errorf("failed to update history for global forwarding rule %s: %w", rule.Name, err)
				}
			}
		}

		return nil
	})
}

// deleteGlobalForwardingRuleRelations deletes all related records for a global forwarding rule.
func (s *Service) deleteGlobalForwardingRuleRelations(tx *gorm.DB, globalForwardingRuleResourceID string) error {
	if err := tx.Where("global_forwarding_rule_resource_id = ?", globalForwardingRuleResourceID).Delete(&bronze.GCPComputeGlobalForwardingRuleLabel{}).Error; err != nil {
		return err
	}
	return nil
}

// createGlobalForwardingRuleRelations creates all related records for a global forwarding rule.
func (s *Service) createGlobalForwardingRuleRelations(tx *gorm.DB, globalForwardingRuleResourceID string, rule *bronze.GCPComputeGlobalForwardingRule) error {
	for i := range rule.Labels {
		rule.Labels[i].GlobalForwardingRuleResourceID = globalForwardingRuleResourceID
	}
	if len(rule.Labels) > 0 {
		if err := tx.Create(&rule.Labels).Error; err != nil {
			return fmt.Errorf("failed to create labels: %w", err)
		}
	}
	return nil
}

// DeleteStaleGlobalForwardingRules removes global forwarding rules that were not collected in the latest run.
// Also closes history records for deleted global forwarding rules.
func (s *Service) DeleteStaleGlobalForwardingRules(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale global forwarding rules
		var staleRules []bronze.GCPComputeGlobalForwardingRule
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleRules).Error; err != nil {
			return err
		}

		// Close history and delete each stale global forwarding rule
		for _, r := range staleRules {
			// Close history
			if err := s.history.CloseHistory(tx, r.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for global forwarding rule %s: %w", r.ResourceID, err)
			}

			// Delete relations
			if err := s.deleteGlobalForwardingRuleRelations(tx, r.ResourceID); err != nil {
				return fmt.Errorf("failed to delete relations for global forwarding rule %s: %w", r.ResourceID, err)
			}

			// Delete global forwarding rule
			if err := tx.Delete(&r).Error; err != nil {
				return fmt.Errorf("failed to delete global forwarding rule %s: %w", r.ResourceID, err)
			}
		}

		return nil
	})
}
