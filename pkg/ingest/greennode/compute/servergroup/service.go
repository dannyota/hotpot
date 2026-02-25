package servergroup

import (
	"context"
	"fmt"
	"time"

	entcompute "github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute/bronzegreennodecomputeservergroup"
	"github.com/dannyota/hotpot/pkg/storage/ent/greennode/compute/bronzegreennodecomputeservergroupmember"
)

// Service handles GreenNode server group ingestion.
type Service struct {
	client    *Client
	entClient *entcompute.Client
	history   *HistoryService
}

// NewService creates a new server group ingestion service.
func NewService(client *Client, entClient *entcompute.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of server group ingestion.
type IngestResult struct {
	GroupCount      int
	CollectedAt     time.Time
	DurationMillis  int64
}

// Ingest fetches server groups from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	groups, err := s.client.ListServerGroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("list server groups: %w", err)
	}

	dataList := make([]*ServerGroupData, 0, len(groups))
	for _, sg := range groups {
		dataList = append(dataList, ConvertServerGroup(sg, projectID, region, collectedAt))
	}

	if err := s.saveServerGroups(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save server groups: %w", err)
	}

	return &IngestResult{
		GroupCount:     len(dataList),
		CollectedAt:   collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveServerGroups(ctx context.Context, groups []*ServerGroupData) error {
	if len(groups) == 0 {
		return nil
	}

	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	for _, data := range groups {
		existing, err := tx.BronzeGreenNodeComputeServerGroup.Query().
			Where(bronzegreennodecomputeservergroup.ID(data.ID)).
			WithMembers().
			First(ctx)
		if err != nil && !entcompute.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing server group %s: %w", data.Name, err)
		}

		diff := DiffServerGroupData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeComputeServerGroup.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for server group %s: %w", data.Name, err)
			}
			continue
		}

		if existing != nil {
			if err := s.deleteGroupChildren(ctx, tx, data.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("delete children for server group %s: %w", data.Name, err)
			}
		}

		var savedGroup *entcompute.BronzeGreenNodeComputeServerGroup
		if existing == nil {
			savedGroup, err = tx.BronzeGreenNodeComputeServerGroup.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetPolicyID(data.PolicyID).
				SetPolicyName(data.PolicyName).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create server group %s: %w", data.Name, err)
			}
		} else {
			savedGroup, err = tx.BronzeGreenNodeComputeServerGroup.UpdateOneID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetPolicyID(data.PolicyID).
				SetPolicyName(data.PolicyName).
				SetRegion(data.Region).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update server group %s: %w", data.Name, err)
			}
		}

		if err := s.createGroupChildren(ctx, tx, savedGroup, data); err != nil {
			tx.Rollback()
			return fmt.Errorf("create children for server group %s: %w", data.Name, err)
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for server group %s: %w", data.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for server group %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func (s *Service) deleteGroupChildren(ctx context.Context, tx *entcompute.Tx, groupID string) error {
	_, err := tx.BronzeGreenNodeComputeServerGroupMember.Delete().
		Where(bronzegreennodecomputeservergroupmember.HasServerGroupWith(bronzegreennodecomputeservergroup.ID(groupID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete members: %w", err)
	}
	return nil
}

func (s *Service) createGroupChildren(ctx context.Context, tx *entcompute.Tx, group *entcompute.BronzeGreenNodeComputeServerGroup, data *ServerGroupData) error {
	for _, m := range data.Members {
		_, err := tx.BronzeGreenNodeComputeServerGroupMember.Create().
			SetServerGroup(group).
			SetUUID(m.UUID).
			SetName(m.Name).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create member %s: %w", m.Name, err)
		}
	}
	return nil
}

// DeleteStaleServerGroups removes server groups not collected in the latest run for the given region.
func (s *Service) DeleteStaleServerGroups(ctx context.Context, projectID, region string, collectedAt time.Time) error {
	now := time.Now()

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("start transaction: %w", err)
	}
	defer func() {
		if v := recover(); v != nil {
			tx.Rollback()
			panic(v)
		}
	}()

	stale, err := tx.BronzeGreenNodeComputeServerGroup.Query().
		Where(
			bronzegreennodecomputeservergroup.ProjectID(projectID),
			bronzegreennodecomputeservergroup.Region(region),
			bronzegreennodecomputeservergroup.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale server groups: %w", err)
	}

	for _, sg := range stale {
		if err := s.history.CloseHistory(ctx, tx, sg.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for server group %s: %w", sg.ID, err)
		}
		if err := s.deleteGroupChildren(ctx, tx, sg.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete children for server group %s: %w", sg.ID, err)
		}
		if err := tx.BronzeGreenNodeComputeServerGroup.DeleteOneID(sg.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete server group %s: %w", sg.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}
