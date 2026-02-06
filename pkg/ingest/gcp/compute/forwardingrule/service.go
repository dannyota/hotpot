package forwardingrule

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute forwarding rule ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new forwarding rule ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for forwarding rule ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of forwarding rule ingestion.
type IngestResult struct {
	ProjectID          string
	ForwardingRuleCount int
	CollectedAt        time.Time
	DurationMillis     int64
}

// Ingest fetches forwarding rules from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch forwarding rules from GCP
	forwardingRules, err := s.client.ListForwardingRules(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list forwarding rules: %w", err)
	}

	// Convert to bronze models
	bronzeRules := make([]bronze.GCPComputeForwardingRule, 0, len(forwardingRules))
	for _, fr := range forwardingRules {
		rule, err := ConvertForwardingRule(fr, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert forwarding rule: %w", err)
		}
		bronzeRules = append(bronzeRules, rule)
	}

	// Save to database
	if err := s.saveForwardingRules(ctx, bronzeRules); err != nil {
		return nil, fmt.Errorf("failed to save forwarding rules: %w", err)
	}

	return &IngestResult{
		ProjectID:          params.ProjectID,
		ForwardingRuleCount: len(bronzeRules),
		CollectedAt:        collectedAt,
		DurationMillis:     time.Since(startTime).Milliseconds(),
	}, nil
}

// saveForwardingRules saves forwarding rules to the database with history tracking.
func (s *Service) saveForwardingRules(ctx context.Context, forwardingRules []bronze.GCPComputeForwardingRule) error {
	if len(forwardingRules) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, rule := range forwardingRules {
			// Load existing forwarding rule with all relations
			var existing *bronze.GCPComputeForwardingRule
			var old bronze.GCPComputeForwardingRule
			err := tx.Preload("Labels").
				Where("resource_id = ?", rule.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing forwarding rule %s: %w", rule.Name, err)
			}

			// Compute diff
			diff := DiffForwardingRule(existing, &rule)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeForwardingRule{}).
					Where("resource_id = ?", rule.ResourceID).
					Update("collected_at", rule.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for forwarding rule %s: %w", rule.Name, err)
				}
				continue
			}

			// Delete old relations (manual cascade)
			if existing != nil {
				if err := s.deleteForwardingRuleRelations(tx, rule.ResourceID); err != nil {
					return fmt.Errorf("failed to delete old relations for forwarding rule %s: %w", rule.Name, err)
				}
			}

			// Upsert forwarding rule
			if err := tx.Save(&rule).Error; err != nil {
				return fmt.Errorf("failed to upsert forwarding rule %s: %w", rule.Name, err)
			}

			// Create new relations
			if err := s.createForwardingRuleRelations(tx, rule.ResourceID, &rule); err != nil {
				return fmt.Errorf("failed to create relations for forwarding rule %s: %w", rule.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &rule, now); err != nil {
					return fmt.Errorf("failed to create history for forwarding rule %s: %w", rule.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &rule, diff, now); err != nil {
					return fmt.Errorf("failed to update history for forwarding rule %s: %w", rule.Name, err)
				}
			}
		}

		return nil
	})
}

// deleteForwardingRuleRelations deletes all related records for a forwarding rule.
func (s *Service) deleteForwardingRuleRelations(tx *gorm.DB, forwardingRuleResourceID string) error {
	if err := tx.Where("forwarding_rule_resource_id = ?", forwardingRuleResourceID).Delete(&bronze.GCPComputeForwardingRuleLabel{}).Error; err != nil {
		return err
	}
	return nil
}

// createForwardingRuleRelations creates all related records for a forwarding rule.
func (s *Service) createForwardingRuleRelations(tx *gorm.DB, forwardingRuleResourceID string, rule *bronze.GCPComputeForwardingRule) error {
	for i := range rule.Labels {
		rule.Labels[i].ForwardingRuleResourceID = forwardingRuleResourceID
	}
	if len(rule.Labels) > 0 {
		if err := tx.Create(&rule.Labels).Error; err != nil {
			return fmt.Errorf("failed to create labels: %w", err)
		}
	}
	return nil
}

// DeleteStaleForwardingRules removes forwarding rules that were not collected in the latest run.
// Also closes history records for deleted forwarding rules.
func (s *Service) DeleteStaleForwardingRules(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale forwarding rules
		var staleRules []bronze.GCPComputeForwardingRule
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleRules).Error; err != nil {
			return err
		}

		// Close history and delete each stale forwarding rule
		for _, r := range staleRules {
			// Close history
			if err := s.history.CloseHistory(tx, r.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for forwarding rule %s: %w", r.ResourceID, err)
			}

			// Delete relations
			if err := s.deleteForwardingRuleRelations(tx, r.ResourceID); err != nil {
				return fmt.Errorf("failed to delete relations for forwarding rule %s: %w", r.ResourceID, err)
			}

			// Delete forwarding rule
			if err := tx.Delete(&r).Error; err != nil {
				return fmt.Errorf("failed to delete forwarding rule %s: %w", r.ResourceID, err)
			}
		}

		return nil
	})
}
