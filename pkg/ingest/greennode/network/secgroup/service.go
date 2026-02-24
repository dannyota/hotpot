package secgroup

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodenetworksecgroup"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegreennodenetworksecgrouprule"
)

// Service handles GreenNode security group ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new security group ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of security group ingestion.
type IngestResult struct {
	SecgroupCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches security groups from GreenNode and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, projectID, region string) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	secgroups, err := s.client.ListSecgroups(ctx)
	if err != nil {
		return nil, fmt.Errorf("list secgroups: %w", err)
	}

	dataList := make([]*SecgroupData, 0, len(secgroups))
	for _, sg := range secgroups {
		rules, err := s.client.ListSecgroupRulesBySecgroupID(ctx, sg.ID)
		if err != nil {
			return nil, fmt.Errorf("list rules for secgroup %s: %w", sg.ID, err)
		}
		dataList = append(dataList, ConvertSecgroup(sg, rules, projectID, region, collectedAt))
	}

	if err := s.saveSecgroups(ctx, dataList); err != nil {
		return nil, fmt.Errorf("save secgroups: %w", err)
	}

	return &IngestResult{
		SecgroupCount:  len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveSecgroups(ctx context.Context, secgroups []*SecgroupData) error {
	if len(secgroups) == 0 {
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

	for _, data := range secgroups {
		existing, err := tx.BronzeGreenNodeNetworkSecgroup.Query().
			Where(bronzegreennodenetworksecgroup.ID(data.ID)).
			WithRules().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing secgroup %s: %w", data.Name, err)
		}

		diff := DiffSecgroupData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGreenNodeNetworkSecgroup.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for secgroup %s: %w", data.Name, err)
			}
			continue
		}

		if existing != nil {
			if err := s.deleteSecgroupChildren(ctx, tx, data.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("delete children for secgroup %s: %w", data.Name, err)
			}
		}

		var savedSecgroup *ent.BronzeGreenNodeNetworkSecgroup
		if existing == nil {
			savedSecgroup, err = tx.BronzeGreenNodeNetworkSecgroup.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetStatus(data.Status).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create secgroup %s: %w", data.Name, err)
			}
		} else {
			savedSecgroup, err = tx.BronzeGreenNodeNetworkSecgroup.UpdateOneID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetStatus(data.Status).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update secgroup %s: %w", data.Name, err)
			}
		}

		if err := s.createSecgroupChildren(ctx, tx, savedSecgroup, data); err != nil {
			tx.Rollback()
			return fmt.Errorf("create children for secgroup %s: %w", data.Name, err)
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for secgroup %s: %w", data.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for secgroup %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (s *Service) deleteSecgroupChildren(ctx context.Context, tx *ent.Tx, secgroupID string) error {
	_, err := tx.BronzeGreenNodeNetworkSecgroupRule.Delete().
		Where(bronzegreennodenetworksecgrouprule.HasSecgroupWith(bronzegreennodenetworksecgroup.ID(secgroupID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("delete rules: %w", err)
	}
	return nil
}

func (s *Service) createSecgroupChildren(ctx context.Context, tx *ent.Tx, sg *ent.BronzeGreenNodeNetworkSecgroup, data *SecgroupData) error {
	for _, r := range data.Rules {
		_, err := tx.BronzeGreenNodeNetworkSecgroupRule.Create().
			SetSecgroup(sg).
			SetRuleID(r.RuleID).
			SetDirection(r.Direction).
			SetEtherType(r.EtherType).
			SetProtocol(r.Protocol).
			SetDescription(r.Description).
			SetRemoteIPPrefix(r.RemoteIPPrefix).
			SetPortRangeMax(r.PortRangeMax).
			SetPortRangeMin(r.PortRangeMin).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("create rule %s: %w", r.RuleID, err)
		}
	}
	return nil
}

// DeleteStaleSecgroups removes security groups not collected in the latest run for the given region.
func (s *Service) DeleteStaleSecgroups(ctx context.Context, projectID, region string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGreenNodeNetworkSecgroup.Query().
		Where(
			bronzegreennodenetworksecgroup.ProjectID(projectID),
			bronzegreennodenetworksecgroup.Region(region),
			bronzegreennodenetworksecgroup.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale secgroups: %w", err)
	}

	for _, sg := range stale {
		if err := s.history.CloseHistory(ctx, tx, sg.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for secgroup %s: %w", sg.ID, err)
		}
		if err := s.deleteSecgroupChildren(ctx, tx, sg.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete children for secgroup %s: %w", sg.ID, err)
		}
		if err := tx.BronzeGreenNodeNetworkSecgroup.DeleteOneID(sg.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete secgroup %s: %w", sg.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
