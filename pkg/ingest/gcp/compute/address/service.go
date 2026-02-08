package address

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpcomputeaddress"
	"hotpot/pkg/storage/ent/bronzegcpcomputeaddresslabel"
)

// Service handles GCP Compute address ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new address ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for address ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of address ingestion.
type IngestResult struct {
	ProjectID      string
	AddressCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches addresses from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch addresses from GCP
	addresses, err := s.client.ListAddresses(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list addresses: %w", err)
	}

	// Convert to data structs
	addressDataList := make([]*AddressData, 0, len(addresses))
	for _, a := range addresses {
		data, err := ConvertAddress(a, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert address: %w", err)
		}
		addressDataList = append(addressDataList, data)
	}

	// Save to database
	if err := s.saveAddresses(ctx, addressDataList); err != nil {
		return nil, fmt.Errorf("failed to save addresses: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		AddressCount:   len(addressDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveAddresses saves addresses to the database with history tracking.
func (s *Service) saveAddresses(ctx context.Context, addresses []*AddressData) error {
	if len(addresses) == 0 {
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

	for _, addressData := range addresses {
		// Load existing address with all edges
		existing, err := tx.BronzeGCPComputeAddress.Query().
			Where(bronzegcpcomputeaddress.ID(addressData.ID)).
			WithLabels().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing address %s: %w", addressData.Name, err)
		}

		// Compute diff
		diff := DiffAddressData(existing, addressData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeAddress.UpdateOneID(addressData.ID).
				SetCollectedAt(addressData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for address %s: %w", addressData.Name, err)
			}
			continue
		}

		// Delete old child entities if updating
		if existing != nil {
			if err := s.deleteAddressChildren(ctx, tx, addressData.ID); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old children for address %s: %w", addressData.Name, err)
			}
		}

		// Create or update address
		var savedAddress *ent.BronzeGCPComputeAddress
		if existing == nil {
			// Create new address
			create := tx.BronzeGCPComputeAddress.Create().
				SetID(addressData.ID).
				SetName(addressData.Name).
				SetDescription(addressData.Description).
				SetAddress(addressData.Address).
				SetAddressType(addressData.AddressType).
				SetIPVersion(addressData.IpVersion).
				SetIpv6EndpointType(addressData.Ipv6EndpointType).
				SetIPCollection(addressData.IpCollection).
				SetRegion(addressData.Region).
				SetStatus(addressData.Status).
				SetPurpose(addressData.Purpose).
				SetNetwork(addressData.Network).
				SetSubnetwork(addressData.Subnetwork).
				SetNetworkTier(addressData.NetworkTier).
				SetPrefixLength(addressData.PrefixLength).
				SetSelfLink(addressData.SelfLink).
				SetCreationTimestamp(addressData.CreationTimestamp).
				SetLabelFingerprint(addressData.LabelFingerprint).
				SetProjectID(addressData.ProjectID).
				SetCollectedAt(addressData.CollectedAt)

			if addressData.UsersJSON != nil {
				create.SetUsersJSON(addressData.UsersJSON)
			}

			savedAddress, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create address %s: %w", addressData.Name, err)
			}
		} else {
			// Update existing address
			update := tx.BronzeGCPComputeAddress.UpdateOneID(addressData.ID).
				SetName(addressData.Name).
				SetDescription(addressData.Description).
				SetAddress(addressData.Address).
				SetAddressType(addressData.AddressType).
				SetIPVersion(addressData.IpVersion).
				SetIpv6EndpointType(addressData.Ipv6EndpointType).
				SetIPCollection(addressData.IpCollection).
				SetRegion(addressData.Region).
				SetStatus(addressData.Status).
				SetPurpose(addressData.Purpose).
				SetNetwork(addressData.Network).
				SetSubnetwork(addressData.Subnetwork).
				SetNetworkTier(addressData.NetworkTier).
				SetPrefixLength(addressData.PrefixLength).
				SetSelfLink(addressData.SelfLink).
				SetCreationTimestamp(addressData.CreationTimestamp).
				SetLabelFingerprint(addressData.LabelFingerprint).
				SetProjectID(addressData.ProjectID).
				SetCollectedAt(addressData.CollectedAt)

			if addressData.UsersJSON != nil {
				update.SetUsersJSON(addressData.UsersJSON)
			}

			savedAddress, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update address %s: %w", addressData.Name, err)
			}
		}

		// Create child entities
		if err := s.createAddressChildren(ctx, tx, savedAddress, addressData); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to create children for address %s: %w", addressData.Name, err)
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, addressData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for address %s: %w", addressData.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, addressData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for address %s: %w", addressData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// deleteAddressChildren deletes all child entities for an address.
func (s *Service) deleteAddressChildren(ctx context.Context, tx *ent.Tx, addressID string) error {
	// Delete labels
	_, err := tx.BronzeGCPComputeAddressLabel.Delete().
		Where(bronzegcpcomputeaddresslabel.HasAddressWith(bronzegcpcomputeaddress.ID(addressID))).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete labels: %w", err)
	}

	return nil
}

// createAddressChildren creates all child entities for an address.
func (s *Service) createAddressChildren(ctx context.Context, tx *ent.Tx, address *ent.BronzeGCPComputeAddress, data *AddressData) error {
	// Create labels
	for _, labelData := range data.Labels {
		_, err := tx.BronzeGCPComputeAddressLabel.Create().
			SetAddress(address).
			SetKey(labelData.Key).
			SetValue(labelData.Value).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create label %s: %w", labelData.Key, err)
		}
	}

	return nil
}

// DeleteStaleAddresses removes addresses that were not collected in the latest run.
// Also closes history records for deleted addresses.
func (s *Service) DeleteStaleAddresses(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale addresses
	staleAddresses, err := tx.BronzeGCPComputeAddress.Query().
		Where(
			bronzegcpcomputeaddress.ProjectID(projectID),
			bronzegcpcomputeaddress.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to query stale addresses: %w", err)
	}

	// Close history and delete each stale address
	for _, addr := range staleAddresses {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, addr.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for address %s: %w", addr.ID, err)
		}

		// Delete children
		if err := s.deleteAddressChildren(ctx, tx, addr.ID); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete children for address %s: %w", addr.ID, err)
		}

		// Delete address
		if err := tx.BronzeGCPComputeAddress.DeleteOneID(addr.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete address %s: %w", addr.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
