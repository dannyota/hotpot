package network

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpcomputenetwork"
	"hotpot/pkg/storage/ent/bronzegcpcomputenetworkpeering"
)

// Service handles GCP Compute network ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new network ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for network ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of network ingestion.
type IngestResult struct {
	ProjectID      string
	NetworkCount   int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches networks from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch networks from GCP
	networks, err := s.client.ListNetworks(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}

	// Convert to data structs
	networkDataList := make([]*NetworkData, 0, len(networks))
	for _, n := range networks {
		data, err := ConvertNetwork(n, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert network: %w", err)
		}
		networkDataList = append(networkDataList, data)
	}

	// Save to database
	if err := s.saveNetworks(ctx, networkDataList); err != nil {
		return nil, fmt.Errorf("failed to save networks: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		NetworkCount:   len(networkDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveNetworks saves networks to the database with history tracking.
func (s *Service) saveNetworks(ctx context.Context, networks []*NetworkData) error {
	if len(networks) == 0 {
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

	for _, networkData := range networks {
		// Load existing network with peerings
		existing, err := tx.BronzeGCPComputeNetwork.Query().
			Where(bronzegcpcomputenetwork.ID(networkData.ID)).
			WithPeerings().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing network %s: %w", networkData.Name, err)
		}

		// Compute diff
		diff := DiffNetworkData(existing, networkData)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeNetwork.UpdateOneID(networkData.ID).
				SetCollectedAt(networkData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for network %s: %w", networkData.Name, err)
			}
			continue
		}

		// Delete old peerings if updating
		if existing != nil {
			_, err := tx.BronzeGCPComputeNetworkPeering.Delete().
				Where(bronzegcpcomputenetworkpeering.HasNetworkRefWith(bronzegcpcomputenetwork.ID(networkData.ID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old peerings for network %s: %w", networkData.Name, err)
			}
		}

		// Create or update network
		var savedNetwork *ent.BronzeGCPComputeNetwork
		if existing == nil {
			// Create new network
			create := tx.BronzeGCPComputeNetwork.Create().
				SetID(networkData.ID).
				SetName(networkData.Name).
				SetDescription(networkData.Description).
				SetSelfLink(networkData.SelfLink).
				SetCreationTimestamp(networkData.CreationTimestamp).
				SetAutoCreateSubnetworks(networkData.AutoCreateSubnetworks).
				SetMtu(networkData.Mtu).
				SetRoutingMode(networkData.RoutingMode).
				SetNetworkFirewallPolicyEnforcementOrder(networkData.NetworkFirewallPolicyEnforcementOrder).
				SetEnableUlaInternalIpv6(networkData.EnableUlaInternalIpv6).
				SetInternalIpv6Range(networkData.InternalIpv6Range).
				SetGatewayIpv4(networkData.GatewayIpv4).
				SetProjectID(networkData.ProjectID).
				SetCollectedAt(networkData.CollectedAt)

			if networkData.SubnetworksJSON != nil {
				create.SetSubnetworksJSON(networkData.SubnetworksJSON)
			}

			savedNetwork, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create network %s: %w", networkData.Name, err)
			}
		} else {
			// Update existing network
			update := tx.BronzeGCPComputeNetwork.UpdateOneID(networkData.ID).
				SetName(networkData.Name).
				SetDescription(networkData.Description).
				SetSelfLink(networkData.SelfLink).
				SetCreationTimestamp(networkData.CreationTimestamp).
				SetAutoCreateSubnetworks(networkData.AutoCreateSubnetworks).
				SetMtu(networkData.Mtu).
				SetRoutingMode(networkData.RoutingMode).
				SetNetworkFirewallPolicyEnforcementOrder(networkData.NetworkFirewallPolicyEnforcementOrder).
				SetEnableUlaInternalIpv6(networkData.EnableUlaInternalIpv6).
				SetInternalIpv6Range(networkData.InternalIpv6Range).
				SetGatewayIpv4(networkData.GatewayIpv4).
				SetProjectID(networkData.ProjectID).
				SetCollectedAt(networkData.CollectedAt)

			if networkData.SubnetworksJSON != nil {
				update.SetSubnetworksJSON(networkData.SubnetworksJSON)
			}

			savedNetwork, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update network %s: %w", networkData.Name, err)
			}
		}

		// Create new peerings
		for _, peering := range networkData.Peerings {
			_, err := tx.BronzeGCPComputeNetworkPeering.Create().
				SetName(peering.Name).
				SetNetwork(peering.Network).
				SetState(peering.State).
				SetStateDetails(peering.StateDetails).
				SetExportCustomRoutes(peering.ExportCustomRoutes).
				SetImportCustomRoutes(peering.ImportCustomRoutes).
				SetExportSubnetRoutesWithPublicIP(peering.ExportSubnetRoutesWithPublicIp).
				SetImportSubnetRoutesWithPublicIP(peering.ImportSubnetRoutesWithPublicIp).
				SetExchangeSubnetRoutes(peering.ExchangeSubnetRoutes).
				SetStackType(peering.StackType).
				SetPeerMtu(peering.PeerMtu).
				SetAutoCreateRoutes(peering.AutoCreateRoutes).
				SetNetworkRef(savedNetwork).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create peering for network %s: %w", networkData.Name, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, networkData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for network %s: %w", networkData.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, networkData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for network %s: %w", networkData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleNetworks removes networks that were not collected in the latest run.
// Also closes history records for deleted networks.
func (s *Service) DeleteStaleNetworks(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale networks
	staleNetworks, err := tx.BronzeGCPComputeNetwork.Query().
		Where(
			bronzegcpcomputenetwork.ProjectID(projectID),
			bronzegcpcomputenetwork.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale network
	for _, network := range staleNetworks {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, network.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for network %s: %w", network.ID, err)
		}

		// Delete network (peerings will be deleted automatically via CASCADE)
		if err := tx.BronzeGCPComputeNetwork.DeleteOne(network).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete network %s: %w", network.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
