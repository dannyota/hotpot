package instance

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzeawsec2instance"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzeawsec2instancetag"
)

// Service handles AWS EC2 instance ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new instance ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for instance ingestion.
type IngestParams struct {
	AccountID string
	Region    string
}

// IngestResult contains the result of instance ingestion.
type IngestResult struct {
	AccountID      string
	Region         string
	InstanceCount  int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches instances from AWS and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch instances from AWS
	instances, err := s.client.ListInstances(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list instances: %w", err)
	}

	// Convert to data structs
	instanceDataList := make([]*InstanceData, 0, len(instances))
	for _, inst := range instances {
		data, err := ConvertInstance(inst, params.AccountID, params.Region, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert instance: %w", err)
		}
		instanceDataList = append(instanceDataList, data)
	}

	// Save to database
	if err := s.saveInstances(ctx, instanceDataList); err != nil {
		return nil, fmt.Errorf("failed to save instances: %w", err)
	}

	return &IngestResult{
		AccountID:      params.AccountID,
		Region:         params.Region,
		InstanceCount:  len(instanceDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveInstances saves instances to the database with history tracking.
func (s *Service) saveInstances(ctx context.Context, instances []*InstanceData) error {
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
		// Load existing instance with tags
		existing, err := tx.BronzeAWSEC2Instance.Query().
			Where(bronzeawsec2instance.ID(instanceData.ResourceID)).
			WithTags().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing instance %s: %w", instanceData.ResourceID, err)
		}

		// Compute diff
		diff := DiffInstanceData(existing, instanceData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeAWSEC2Instance.UpdateOneID(instanceData.ResourceID).
				SetCollectedAt(instanceData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for instance %s: %w", instanceData.ResourceID, err)
			}
			continue
		}

		// Delete old tags if updating
		if existing != nil {
			if err := s.deleteInstanceChildren(ctx, tx, instanceData.ResourceID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old children for instance %s: %w", instanceData.ResourceID, err)
			}
		}

		// Create or update instance
		var savedInstance *ent.BronzeAWSEC2Instance
		if existing == nil {
			create := tx.BronzeAWSEC2Instance.Create().
				SetID(instanceData.ResourceID).
				SetName(instanceData.Name).
				SetInstanceType(instanceData.InstanceType).
				SetState(instanceData.State).
				SetVpcID(instanceData.VpcID).
				SetSubnetID(instanceData.SubnetID).
				SetPrivateIPAddress(instanceData.PrivateIPAddress).
				SetPublicIPAddress(instanceData.PublicIPAddress).
				SetAmiID(instanceData.AmiID).
				SetKeyName(instanceData.KeyName).
				SetPlatform(instanceData.Platform).
				SetArchitecture(instanceData.Architecture).
				SetAccountID(instanceData.AccountID).
				SetRegion(instanceData.Region).
				SetCollectedAt(instanceData.CollectedAt).
				SetFirstCollectedAt(instanceData.CollectedAt)

			if instanceData.LaunchTime != nil {
				create.SetLaunchTime(*instanceData.LaunchTime)
			}
			if instanceData.SecurityGroupJSON != nil {
				create.SetSecurityGroupsJSON(instanceData.SecurityGroupJSON)
			}

			savedInstance, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create instance %s: %w", instanceData.ResourceID, err)
			}
		} else {
			update := tx.BronzeAWSEC2Instance.UpdateOneID(instanceData.ResourceID).
				SetName(instanceData.Name).
				SetInstanceType(instanceData.InstanceType).
				SetState(instanceData.State).
				SetVpcID(instanceData.VpcID).
				SetSubnetID(instanceData.SubnetID).
				SetPrivateIPAddress(instanceData.PrivateIPAddress).
				SetPublicIPAddress(instanceData.PublicIPAddress).
				SetAmiID(instanceData.AmiID).
				SetKeyName(instanceData.KeyName).
				SetPlatform(instanceData.Platform).
				SetArchitecture(instanceData.Architecture).
				SetAccountID(instanceData.AccountID).
				SetRegion(instanceData.Region).
				SetCollectedAt(instanceData.CollectedAt)

			if instanceData.LaunchTime != nil {
				update.SetLaunchTime(*instanceData.LaunchTime)
			}
			if instanceData.SecurityGroupJSON != nil {
				update.SetSecurityGroupsJSON(instanceData.SecurityGroupJSON)
			}

			savedInstance, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update instance %s: %w", instanceData.ResourceID, err)
			}
		}

		// Create tags
		if err := s.createInstanceChildren(ctx, tx, savedInstance, instanceData); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create children for instance %s: %w", instanceData.ResourceID, err)
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, instanceData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for instance %s: %w", instanceData.ResourceID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, instanceData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for instance %s: %w", instanceData.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Service) deleteInstanceChildren(ctx context.Context, tx *ent.Tx, instanceID string) error {
	_, err := tx.BronzeAWSEC2InstanceTag.Delete().
		Where(bronzeawsec2instancetag.HasInstanceWith(bronzeawsec2instance.ID(instanceID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete tags: %w", err)
	}
	return nil
}

func (s *Service) createInstanceChildren(ctx context.Context, tx *ent.Tx, instance *ent.BronzeAWSEC2Instance, data *InstanceData) error {
	for _, tagData := range data.Tags {
		_, err := tx.BronzeAWSEC2InstanceTag.Create().
			SetInstance(instance).
			SetKey(tagData.Key).
			SetValue(tagData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create tag: %w", err)
		}
	}
	return nil
}

// DeleteStaleInstances removes instances that were not collected in the latest run.
func (s *Service) DeleteStaleInstances(ctx context.Context, accountID, region string, collectedAt time.Time) error {
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

	staleInstances, err := tx.BronzeAWSEC2Instance.Query().
		Where(
			bronzeawsec2instance.AccountID(accountID),
			bronzeawsec2instance.Region(region),
			bronzeawsec2instance.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, inst := range staleInstances {
		if err := s.history.CloseHistory(ctx, tx, inst.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for instance %s: %w", inst.ID, err)
		}

		if err := s.deleteInstanceChildren(ctx, tx, inst.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete children for instance %s: %w", inst.ID, err)
		}

		if err := tx.BronzeAWSEC2Instance.DeleteOne(inst).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete instance %s: %w", inst.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
