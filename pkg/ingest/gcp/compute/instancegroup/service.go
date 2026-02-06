package instancegroup

import (
	"context"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"hotpot/pkg/base/models/bronze"
)

// Service handles GCP Compute instance group ingestion.
type Service struct {
	client  *Client
	db      *gorm.DB
	history *HistoryService
}

// NewService creates a new instance group ingestion service.
func NewService(client *Client, db *gorm.DB) *Service {
	return &Service{
		client:  client,
		db:      db,
		history: NewHistoryService(db),
	}
}

// IngestParams contains parameters for instance group ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of instance group ingestion.
type IngestResult struct {
	ProjectID          string
	InstanceGroupCount int
	CollectedAt        time.Time
	DurationMillis     int64
}

// Ingest fetches instance groups from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch instance groups from GCP
	groups, err := s.client.ListInstanceGroups(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list instance groups: %w", err)
	}

	// For each group, fetch members and convert
	bronzeGroups := make([]bronze.GCPComputeInstanceGroup, 0, len(groups))
	for _, g := range groups {
		// Extract zone name from zone URL for the members API call
		zoneName := extractZoneName(g.GetZone())

		members, err := s.client.ListInstanceGroupMembers(ctx, params.ProjectID, zoneName, g.GetName())
		if err != nil {
			return nil, fmt.Errorf("failed to list members for instance group %s: %w", g.GetName(), err)
		}

		bronzeGroups = append(bronzeGroups, ConvertInstanceGroup(g, members, params.ProjectID, collectedAt))
	}

	// Save to database
	if err := s.saveInstanceGroups(ctx, bronzeGroups); err != nil {
		return nil, fmt.Errorf("failed to save instance groups: %w", err)
	}

	return &IngestResult{
		ProjectID:          params.ProjectID,
		InstanceGroupCount: len(bronzeGroups),
		CollectedAt:        collectedAt,
		DurationMillis:     time.Since(startTime).Milliseconds(),
	}, nil
}

// extractZoneName extracts the zone name from a full zone URL.
// Example: "https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-a" -> "us-central1-a"
func extractZoneName(zoneURL string) string {
	parts := strings.Split(zoneURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return zoneURL
}

// saveInstanceGroups saves instance groups to the database with history tracking.
func (s *Service) saveInstanceGroups(ctx context.Context, groups []bronze.GCPComputeInstanceGroup) error {
	if len(groups) == 0 {
		return nil
	}

	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, group := range groups {
			// Load existing group with all relations
			var existing *bronze.GCPComputeInstanceGroup
			var old bronze.GCPComputeInstanceGroup
			err := tx.Preload("NamedPorts").Preload("Members").
				Where("resource_id = ?", group.ResourceID).
				First(&old).Error
			if err == nil {
				existing = &old
			} else if err != gorm.ErrRecordNotFound {
				return fmt.Errorf("failed to load existing instance group %s: %w", group.Name, err)
			}

			// Compute diff
			diff := DiffInstanceGroup(existing, &group)

			// Skip if no changes
			if !diff.HasAnyChange() && existing != nil {
				// Update collected_at only
				if err := tx.Model(&bronze.GCPComputeInstanceGroup{}).
					Where("resource_id = ?", group.ResourceID).
					Update("collected_at", group.CollectedAt).Error; err != nil {
					return fmt.Errorf("failed to update collected_at for instance group %s: %w", group.Name, err)
				}
				continue
			}

			// Delete old relations (manual cascade)
			if existing != nil {
				if err := s.deleteInstanceGroupRelations(tx, group.ResourceID); err != nil {
					return fmt.Errorf("failed to delete old relations for instance group %s: %w", group.Name, err)
				}
			}

			// Upsert instance group
			if err := tx.Save(&group).Error; err != nil {
				return fmt.Errorf("failed to upsert instance group %s: %w", group.Name, err)
			}

			// Create new relations
			if err := s.createInstanceGroupRelations(tx, group.ResourceID, &group); err != nil {
				return fmt.Errorf("failed to create relations for instance group %s: %w", group.Name, err)
			}

			// Track history
			if diff.IsNew {
				if err := s.history.CreateHistory(tx, &group, now); err != nil {
					return fmt.Errorf("failed to create history for instance group %s: %w", group.Name, err)
				}
			} else {
				if err := s.history.UpdateHistory(tx, existing, &group, diff, now); err != nil {
					return fmt.Errorf("failed to update history for instance group %s: %w", group.Name, err)
				}
			}
		}

		return nil
	})
}

// deleteInstanceGroupRelations deletes all related records for an instance group.
func (s *Service) deleteInstanceGroupRelations(tx *gorm.DB, groupResourceID string) error {
	// Delete named ports
	if err := tx.Where("group_resource_id = ?", groupResourceID).Delete(&bronze.GCPComputeInstanceGroupNamedPort{}).Error; err != nil {
		return err
	}

	// Delete members
	if err := tx.Where("group_resource_id = ?", groupResourceID).Delete(&bronze.GCPComputeInstanceGroupMember{}).Error; err != nil {
		return err
	}

	return nil
}

// createInstanceGroupRelations creates all related records for an instance group.
func (s *Service) createInstanceGroupRelations(tx *gorm.DB, groupResourceID string, group *bronze.GCPComputeInstanceGroup) error {
	// Create named ports
	for i := range group.NamedPorts {
		group.NamedPorts[i].GroupResourceID = groupResourceID
	}
	if len(group.NamedPorts) > 0 {
		if err := tx.Create(&group.NamedPorts).Error; err != nil {
			return fmt.Errorf("failed to create named ports: %w", err)
		}
	}

	// Create members
	for i := range group.Members {
		group.Members[i].GroupResourceID = groupResourceID
	}
	if len(group.Members) > 0 {
		if err := tx.Create(&group.Members).Error; err != nil {
			return fmt.Errorf("failed to create members: %w", err)
		}
	}

	return nil
}

// DeleteStaleInstanceGroups removes instance groups that were not collected in the latest run.
// Also closes history records for deleted instance groups.
func (s *Service) DeleteStaleInstanceGroups(ctx context.Context, projectID string, collectedAt time.Time) error {
	now := time.Now()

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Find stale instance groups
		var staleGroups []bronze.GCPComputeInstanceGroup
		if err := tx.Where("project_id = ? AND collected_at < ?", projectID, collectedAt).
			Find(&staleGroups).Error; err != nil {
			return err
		}

		// Close history and delete each stale instance group
		for _, g := range staleGroups {
			// Close history
			if err := s.history.CloseHistory(tx, g.ResourceID, now); err != nil {
				return fmt.Errorf("failed to close history for instance group %s: %w", g.ResourceID, err)
			}

			// Delete relations
			if err := s.deleteInstanceGroupRelations(tx, g.ResourceID); err != nil {
				return fmt.Errorf("failed to delete relations for instance group %s: %w", g.ResourceID, err)
			}

			// Delete instance group
			if err := tx.Delete(&g).Error; err != nil {
				return fmt.Errorf("failed to delete instance group %s: %w", g.ResourceID, err)
			}
		}

		return nil
	})
}
