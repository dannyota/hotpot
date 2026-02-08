package subnetwork

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpcomputesubnetwork"
	"hotpot/pkg/storage/ent/bronzegcpcomputesubnetworksecondaryrange"
)

// Service handles GCP Compute subnetwork ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new subnetwork ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for subnetwork ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of subnetwork ingestion.
type IngestResult struct {
	ProjectID       string
	SubnetworkCount int
	CollectedAt     time.Time
	DurationMillis  int64
}

// Ingest fetches subnetworks from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch subnetworks from GCP
	subnetworks, err := s.client.ListSubnetworks(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list subnetworks: %w", err)
	}

	// Convert to data structs
	subnetworkDataList := make([]*SubnetworkData, 0, len(subnetworks))
	for _, sn := range subnetworks {
		data, err := ConvertSubnetwork(sn, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert subnetwork: %w", err)
		}
		subnetworkDataList = append(subnetworkDataList, data)
	}

	// Save to database
	if err := s.saveSubnetworks(ctx, subnetworkDataList); err != nil {
		return nil, fmt.Errorf("failed to save subnetworks: %w", err)
	}

	return &IngestResult{
		ProjectID:       params.ProjectID,
		SubnetworkCount: len(subnetworkDataList),
		CollectedAt:     collectedAt,
		DurationMillis:  time.Since(startTime).Milliseconds(),
	}, nil
}

// saveSubnetworks saves subnetworks to the database with history tracking.
func (s *Service) saveSubnetworks(ctx context.Context, subnetworks []*SubnetworkData) error {
	if len(subnetworks) == 0 {
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

	for _, subnetData := range subnetworks {
		// Load existing subnetwork with secondary ranges
		existing, err := tx.BronzeGCPComputeSubnetwork.Query().
			Where(bronzegcpcomputesubnetwork.ID(subnetData.ID)).
			WithSecondaryIPRanges().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing subnetwork %s: %w", subnetData.Name, err)
		}

		// Compute diff
		diff := DiffSubnetworkData(existing, subnetData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeSubnetwork.UpdateOneID(subnetData.ID).
				SetCollectedAt(subnetData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for subnetwork %s: %w", subnetData.Name, err)
			}
			continue
		}

		// Delete old secondary ranges if updating
		if existing != nil {
			_, err := tx.BronzeGCPComputeSubnetworkSecondaryRange.Delete().
				Where(bronzegcpcomputesubnetworksecondaryrange.HasSubnetworkWith(bronzegcpcomputesubnetwork.ID(subnetData.ID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old secondary ranges for subnetwork %s: %w", subnetData.Name, err)
			}
		}

		// Create or update subnetwork
		var savedSubnet *ent.BronzeGCPComputeSubnetwork
		if existing == nil {
			// Create new subnetwork
			create := tx.BronzeGCPComputeSubnetwork.Create().
				SetID(subnetData.ID).
				SetName(subnetData.Name).
				SetDescription(subnetData.Description).
				SetSelfLink(subnetData.SelfLink).
				SetCreationTimestamp(subnetData.CreationTimestamp).
				SetNetwork(subnetData.Network).
				SetRegion(subnetData.Region).
				SetIPCidrRange(subnetData.IpCidrRange).
				SetGatewayAddress(subnetData.GatewayAddress).
				SetPurpose(subnetData.Purpose).
				SetRole(subnetData.Role).
				SetPrivateIPGoogleAccess(subnetData.PrivateIpGoogleAccess).
				SetPrivateIpv6GoogleAccess(subnetData.PrivateIpv6GoogleAccess).
				SetStackType(subnetData.StackType).
				SetIpv6AccessType(subnetData.Ipv6AccessType).
				SetInternalIpv6Prefix(subnetData.InternalIpv6Prefix).
				SetExternalIpv6Prefix(subnetData.ExternalIpv6Prefix).
				SetFingerprint(subnetData.Fingerprint).
				SetProjectID(subnetData.ProjectID).
				SetCollectedAt(subnetData.CollectedAt).
				SetFirstCollectedAt(subnetData.CollectedAt)

			if subnetData.LogConfigJSON != nil {
				create.SetLogConfigJSON(subnetData.LogConfigJSON)
			}

			savedSubnet, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create subnetwork %s: %w", subnetData.Name, err)
			}
		} else {
			// Update existing subnetwork
			update := tx.BronzeGCPComputeSubnetwork.UpdateOneID(subnetData.ID).
				SetName(subnetData.Name).
				SetDescription(subnetData.Description).
				SetSelfLink(subnetData.SelfLink).
				SetCreationTimestamp(subnetData.CreationTimestamp).
				SetNetwork(subnetData.Network).
				SetRegion(subnetData.Region).
				SetIPCidrRange(subnetData.IpCidrRange).
				SetGatewayAddress(subnetData.GatewayAddress).
				SetPurpose(subnetData.Purpose).
				SetRole(subnetData.Role).
				SetPrivateIPGoogleAccess(subnetData.PrivateIpGoogleAccess).
				SetPrivateIpv6GoogleAccess(subnetData.PrivateIpv6GoogleAccess).
				SetStackType(subnetData.StackType).
				SetIpv6AccessType(subnetData.Ipv6AccessType).
				SetInternalIpv6Prefix(subnetData.InternalIpv6Prefix).
				SetExternalIpv6Prefix(subnetData.ExternalIpv6Prefix).
				SetFingerprint(subnetData.Fingerprint).
				SetProjectID(subnetData.ProjectID).
				SetCollectedAt(subnetData.CollectedAt)

			if subnetData.LogConfigJSON != nil {
				update.SetLogConfigJSON(subnetData.LogConfigJSON)
			}

			savedSubnet, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update subnetwork %s: %w", subnetData.Name, err)
			}
		}

		// Create new secondary ranges
		for _, rangeData := range subnetData.SecondaryIpRanges {
			_, err := tx.BronzeGCPComputeSubnetworkSecondaryRange.Create().
				SetRangeName(rangeData.RangeName).
				SetIPCidrRange(rangeData.IpCidrRange).
				SetSubnetwork(savedSubnet).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create secondary range for subnetwork %s: %w", subnetData.Name, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, subnetData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for subnetwork %s: %w", subnetData.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, subnetData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for subnetwork %s: %w", subnetData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleSubnetworks removes subnetworks that were not collected in the latest run.
// Also closes history records for deleted subnetworks.
func (s *Service) DeleteStaleSubnetworks(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale subnetworks
	staleSubnets, err := tx.BronzeGCPComputeSubnetwork.Query().
		Where(
			bronzegcpcomputesubnetwork.ProjectID(projectID),
			bronzegcpcomputesubnetwork.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale subnetwork
	for _, subnet := range staleSubnets {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, subnet.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for subnetwork %s: %w", subnet.ID, err)
		}

		// Delete subnetwork (secondary ranges will be deleted automatically via CASCADE)
		if err := tx.BronzeGCPComputeSubnetwork.DeleteOne(subnet).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete subnetwork %s: %w", subnet.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
