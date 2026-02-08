package vpngateway

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpvpngateway"
	"hotpot/pkg/storage/ent/bronzegcpvpngatewaylabel"
)

// Service handles GCP Compute VPN gateway ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new VPN gateway ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for VPN gateway ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of VPN gateway ingestion.
type IngestResult struct {
	ProjectID       string
	VpnGatewayCount int
	CollectedAt     time.Time
	DurationMillis  int64
}

// Ingest fetches VPN gateways from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch VPN gateways from GCP
	vpnGateways, err := s.client.ListVpnGateways(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list vpn gateways: %w", err)
	}

	// Convert to data structs
	vpnGatewayDataList := make([]*VpnGatewayData, 0, len(vpnGateways))
	for _, gw := range vpnGateways {
		data, err := ConvertVpnGateway(gw, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert vpn gateway: %w", err)
		}
		vpnGatewayDataList = append(vpnGatewayDataList, data)
	}

	// Save to database
	if err := s.saveVpnGateways(ctx, vpnGatewayDataList); err != nil {
		return nil, fmt.Errorf("failed to save vpn gateways: %w", err)
	}

	return &IngestResult{
		ProjectID:       params.ProjectID,
		VpnGatewayCount: len(vpnGatewayDataList),
		CollectedAt:     collectedAt,
		DurationMillis:  time.Since(startTime).Milliseconds(),
	}, nil
}

// saveVpnGateways saves VPN gateways to the database with history tracking.
func (s *Service) saveVpnGateways(ctx context.Context, vpnGateways []*VpnGatewayData) error {
	if len(vpnGateways) == 0 {
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

	for _, data := range vpnGateways {
		// Load existing VPN gateway with all relations
		existing, err := tx.BronzeGCPVPNGateway.Query().
			Where(bronzegcpvpngateway.ID(data.ID)).
			WithLabels().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("query existing vpn gateway %s: %w", data.Name, err)
		}

		// Compute diff
		diff := DiffVpnGatewayData(existing, data)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			_, err := tx.BronzeGCPVPNGateway.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for vpn gateway %s: %w", data.Name, err)
			}
			continue
		}

		// Upsert VPN gateway
		var savedGateway *ent.BronzeGCPVPNGateway
		if existing == nil {
			// Create new VPN gateway
			create := tx.BronzeGCPVPNGateway.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetRegion(data.Region).
				SetNetwork(data.Network).
				SetSelfLink(data.SelfLink).
				SetCreationTimestamp(data.CreationTimestamp).
				SetLabelFingerprint(data.LabelFingerprint).
				SetGatewayIPVersion(data.GatewayIpVersion).
				SetStackType(data.StackType).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt)

			if len(data.VpnInterfacesJSON) > 0 {
				create.SetVpnInterfacesJSON(data.VpnInterfacesJSON)
			}

			savedGateway, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create vpn gateway %s: %w", data.Name, err)
			}
		} else {
			// Update existing VPN gateway
			update := tx.BronzeGCPVPNGateway.UpdateOneID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetRegion(data.Region).
				SetNetwork(data.Network).
				SetSelfLink(data.SelfLink).
				SetCreationTimestamp(data.CreationTimestamp).
				SetLabelFingerprint(data.LabelFingerprint).
				SetGatewayIPVersion(data.GatewayIpVersion).
				SetStackType(data.StackType).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt)

			if len(data.VpnInterfacesJSON) > 0 {
				update.SetVpnInterfacesJSON(data.VpnInterfacesJSON)
			} else {
				update.ClearVpnInterfacesJSON()
			}

			savedGateway, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update vpn gateway %s: %w", data.Name, err)
			}

			// Delete old labels (cascade handled by edges)
			_, err = tx.BronzeGCPVPNGatewayLabel.Delete().
				Where(bronzegcpvpngatewaylabel.HasVpnGatewayWith(bronzegcpvpngateway.ID(data.ID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("delete old labels for vpn gateway %s: %w", data.Name, err)
			}
		}

		// Create labels
		for _, labelData := range data.Labels {
			_, err := tx.BronzeGCPVPNGatewayLabel.Create().
				SetKey(labelData.Key).
				SetValue(labelData.Value).
				SetVpnGateway(savedGateway).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create label for vpn gateway %s: %w", data.Name, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for vpn gateway %s: %w", data.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for vpn gateway %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleVpnGateways removes VPN gateways that were not collected in the latest run.
// Also closes history records for deleted VPN gateways.
func (s *Service) DeleteStaleVpnGateways(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale VPN gateways
	staleVpnGateways, err := tx.BronzeGCPVPNGateway.Query().
		Where(
			bronzegcpvpngateway.ProjectID(projectID),
			bronzegcpvpngateway.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale vpn gateways: %w", err)
	}

	// Close history and delete each stale VPN gateway
	for _, gw := range staleVpnGateways {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, gw.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for vpn gateway %s: %w", gw.ID, err)
		}

		// Delete VPN gateway (labels cascade via edge)
		if err := tx.BronzeGCPVPNGateway.DeleteOneID(gw.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete vpn gateway %s: %w", gw.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
