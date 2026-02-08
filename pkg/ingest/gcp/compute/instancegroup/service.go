package instancegroup

import (
	"context"
	"fmt"
	"strings"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpcomputeinstancegroup"
	"hotpot/pkg/storage/ent/bronzegcpcomputeinstancegroupmember"
	"hotpot/pkg/storage/ent/bronzegcpcomputeinstancegroupnamedport"
)

// Service handles GCP Compute instance group ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new instance group ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
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
	groupDataList := make([]*InstanceGroupData, 0, len(groups))
	for _, g := range groups {
		// Extract zone name from zone URL for the members API call
		zoneName := extractZoneName(g.GetZone())

		members, err := s.client.ListInstanceGroupMembers(ctx, params.ProjectID, zoneName, g.GetName())
		if err != nil {
			return nil, fmt.Errorf("failed to list members for instance group %s: %w", g.GetName(), err)
		}

		groupDataList = append(groupDataList, ConvertInstanceGroup(g, members, params.ProjectID, collectedAt))
	}

	// Save to database
	if err := s.saveInstanceGroups(ctx, groupDataList); err != nil {
		return nil, fmt.Errorf("failed to save instance groups: %w", err)
	}

	return &IngestResult{
		ProjectID:          params.ProjectID,
		InstanceGroupCount: len(groupDataList),
		CollectedAt:        collectedAt,
		DurationMillis:     time.Since(startTime).Milliseconds(),
	}, nil
}

// extractZoneName extracts the zone name from a zone URL.
// Example: "https://www.googleapis.com/compute/v1/projects/my-project/zones/us-central1-a" -> "us-central1-a"
func extractZoneName(zoneURL string) string {
	parts := strings.Split(zoneURL, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

// saveInstanceGroups saves instance groups to the database with history tracking.
func (s *Service) saveInstanceGroups(ctx context.Context, groups []*InstanceGroupData) error {
	if len(groups) == 0 {
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

	for _, groupData := range groups {
		// Load existing instance group with edges
		existing, err := tx.BronzeGCPComputeInstanceGroup.Query().
			Where(bronzegcpcomputeinstancegroup.ID(groupData.ID)).
			WithNamedPorts().
			WithMembers().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing instance group %s: %w", groupData.ID, err)
		}

		// Compute diff
		diff := DiffInstanceGroupData(existing, groupData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeInstanceGroup.UpdateOneID(groupData.ID).
				SetCollectedAt(groupData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for instance group %s: %w", groupData.ID, err)
			}
			continue
		}

		// Delete old children if updating
		if existing != nil {
			// Delete named ports
			_, err := tx.BronzeGCPComputeInstanceGroupNamedPort.Delete().
				Where(bronzegcpcomputeinstancegroupnamedport.HasInstanceGroupWith(bronzegcpcomputeinstancegroup.ID(groupData.ID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old named ports for instance group %s: %w", groupData.ID, err)
			}

			// Delete members
			_, err = tx.BronzeGCPComputeInstanceGroupMember.Delete().
				Where(bronzegcpcomputeinstancegroupmember.HasInstanceGroupWith(bronzegcpcomputeinstancegroup.ID(groupData.ID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old members for instance group %s: %w", groupData.ID, err)
			}
		}

		// Create or update instance group
		var savedGroup *ent.BronzeGCPComputeInstanceGroup
		if existing == nil {
			// Create new instance group
			savedGroup, err = tx.BronzeGCPComputeInstanceGroup.Create().
				SetID(groupData.ID).
				SetName(groupData.Name).
				SetDescription(groupData.Description).
				SetZone(groupData.Zone).
				SetNetwork(groupData.Network).
				SetSubnetwork(groupData.Subnetwork).
				SetSize(groupData.Size).
				SetSelfLink(groupData.SelfLink).
				SetCreationTimestamp(groupData.CreationTimestamp).
				SetFingerprint(groupData.Fingerprint).
				SetProjectID(groupData.ProjectID).
				SetCollectedAt(groupData.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create instance group %s: %w", groupData.ID, err)
			}
		} else {
			// Update existing instance group
			savedGroup, err = tx.BronzeGCPComputeInstanceGroup.UpdateOneID(groupData.ID).
				SetName(groupData.Name).
				SetDescription(groupData.Description).
				SetZone(groupData.Zone).
				SetNetwork(groupData.Network).
				SetSubnetwork(groupData.Subnetwork).
				SetSize(groupData.Size).
				SetSelfLink(groupData.SelfLink).
				SetCreationTimestamp(groupData.CreationTimestamp).
				SetFingerprint(groupData.Fingerprint).
				SetCollectedAt(groupData.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update instance group %s: %w", groupData.ID, err)
			}
		}

		// Create named ports
		for _, port := range groupData.NamedPorts {
			_, err := tx.BronzeGCPComputeInstanceGroupNamedPort.Create().
				SetName(port.Name).
				SetPort(port.Port).
				SetInstanceGroup(savedGroup).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create named port for instance group %s: %w", groupData.ID, err)
			}
		}

		// Create members
		for _, member := range groupData.Members {
			_, err := tx.BronzeGCPComputeInstanceGroupMember.Create().
				SetInstanceURL(member.InstanceURL).
				SetInstanceName(member.InstanceName).
				SetStatus(member.Status).
				SetInstanceGroup(savedGroup).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create member for instance group %s: %w", groupData.ID, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, groupData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for instance group %s: %w", groupData.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, groupData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for instance group %s: %w", groupData.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleInstanceGroups removes instance groups that were not collected in the latest run.
// Also closes history records for deleted instance groups.
func (s *Service) DeleteStaleInstanceGroups(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale instance groups
	staleGroups, err := tx.BronzeGCPComputeInstanceGroup.Query().
		Where(
			bronzegcpcomputeinstancegroup.ProjectID(projectID),
			bronzegcpcomputeinstancegroup.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale instance group
	for _, group := range staleGroups {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, group.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for instance group %s: %w", group.ID, err)
		}

		// Delete instance group (named ports and members will be deleted automatically via CASCADE)
		if err := tx.BronzeGCPComputeInstanceGroup.DeleteOne(group).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete instance group %s: %w", group.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
