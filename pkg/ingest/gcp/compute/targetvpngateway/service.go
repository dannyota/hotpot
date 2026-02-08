package targetvpngateway

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpvpntargetgateway"
	"hotpot/pkg/storage/ent/bronzegcpvpntargetgatewaylabel"
)

// Service handles GCP Compute Classic VPN gateway ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new target VPN gateway ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for target VPN gateway ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of target VPN gateway ingestion.
type IngestResult struct {
	ProjectID             string
	TargetVpnGatewayCount int
	CollectedAt           time.Time
	DurationMillis        int64
}

// Ingest fetches Classic VPN gateways from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch target VPN gateways from GCP
	targetVpnGateways, err := s.client.ListTargetVpnGateways(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list target vpn gateways: %w", err)
	}

	// Convert to data structs
	targetVpnGatewayDataList := make([]*TargetVpnGatewayData, 0, len(targetVpnGateways))
	for _, gw := range targetVpnGateways {
		data, err := ConvertTargetVpnGateway(gw, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert target vpn gateway: %w", err)
		}
		targetVpnGatewayDataList = append(targetVpnGatewayDataList, data)
	}

	// Save to database
	if err := s.saveTargetVpnGateways(ctx, targetVpnGatewayDataList); err != nil {
		return nil, fmt.Errorf("failed to save target vpn gateways: %w", err)
	}

	return &IngestResult{
		ProjectID:             params.ProjectID,
		TargetVpnGatewayCount: len(targetVpnGatewayDataList),
		CollectedAt:           collectedAt,
		DurationMillis:        time.Since(startTime).Milliseconds(),
	}, nil
}

// saveTargetVpnGateways saves target VPN gateways to the database with history tracking.
func (s *Service) saveTargetVpnGateways(ctx context.Context, targetVpnGateways []*TargetVpnGatewayData) error {
	if len(targetVpnGateways) == 0 {
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

	for _, data := range targetVpnGateways {
		// Load existing target VPN gateway with all relations
		existing, err := tx.BronzeGCPVPNTargetGateway.Query().
			Where(bronzegcpvpntargetgateway.ID(data.ID)).
			WithLabels().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("query existing target vpn gateway %s: %w", data.Name, err)
		}

		// Compute diff
		diff := DiffTargetVpnGatewayData(existing, data)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			_, err := tx.BronzeGCPVPNTargetGateway.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for target vpn gateway %s: %w", data.Name, err)
			}
			continue
		}

		// Upsert target VPN gateway
		var savedGateway *ent.BronzeGCPVPNTargetGateway
		if existing == nil {
			// Create new target VPN gateway
			create := tx.BronzeGCPVPNTargetGateway.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetStatus(data.Status).
				SetRegion(data.Region).
				SetNetwork(data.Network).
				SetSelfLink(data.SelfLink).
				SetCreationTimestamp(data.CreationTimestamp).
				SetLabelFingerprint(data.LabelFingerprint).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt)

			if len(data.ForwardingRulesJSON) > 0 {
				create.SetForwardingRulesJSON(data.ForwardingRulesJSON)
			}

			if len(data.TunnelsJSON) > 0 {
				create.SetTunnelsJSON(data.TunnelsJSON)
			}

			savedGateway, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create target vpn gateway %s: %w", data.Name, err)
			}
		} else {
			// Update existing target VPN gateway
			update := tx.BronzeGCPVPNTargetGateway.UpdateOneID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetStatus(data.Status).
				SetRegion(data.Region).
				SetNetwork(data.Network).
				SetSelfLink(data.SelfLink).
				SetCreationTimestamp(data.CreationTimestamp).
				SetLabelFingerprint(data.LabelFingerprint).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt)

			if len(data.ForwardingRulesJSON) > 0 {
				update.SetForwardingRulesJSON(data.ForwardingRulesJSON)
			} else {
				update.ClearForwardingRulesJSON()
			}

			if len(data.TunnelsJSON) > 0 {
				update.SetTunnelsJSON(data.TunnelsJSON)
			} else {
				update.ClearTunnelsJSON()
			}

			savedGateway, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("update target vpn gateway %s: %w", data.Name, err)
			}

			// Delete old labels (cascade handled by edges)
			_, err = tx.BronzeGCPVPNTargetGatewayLabel.Delete().
				Where(bronzegcpvpntargetgatewaylabel.HasTargetVpnGatewayWith(bronzegcpvpntargetgateway.ID(data.ID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("delete old labels for target vpn gateway %s: %w", data.Name, err)
			}
		}

		// Create labels
		for _, labelData := range data.Labels {
			_, err := tx.BronzeGCPVPNTargetGatewayLabel.Create().
				SetKey(labelData.Key).
				SetValue(labelData.Value).
				SetTargetVpnGateway(savedGateway).
				Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("create label for target vpn gateway %s: %w", data.Name, err)
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for target vpn gateway %s: %w", data.Name, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for target vpn gateway %s: %w", data.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleTargetVpnGateways removes target VPN gateways that were not collected in the latest run.
// Also closes history records for deleted target VPN gateways.
func (s *Service) DeleteStaleTargetVpnGateways(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale target VPN gateways
	staleTargetVpnGateways, err := tx.BronzeGCPVPNTargetGateway.Query().
		Where(
			bronzegcpvpntargetgateway.ProjectID(projectID),
			bronzegcpvpntargetgateway.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("query stale target vpn gateways: %w", err)
	}

	// Close history and delete each stale target VPN gateway
	for _, gw := range staleTargetVpnGateways {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, gw.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for target vpn gateway %s: %w", gw.ID, err)
		}

		// Delete target VPN gateway (labels cascade via edge)
		if err := tx.BronzeGCPVPNTargetGateway.DeleteOneID(gw.ID).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete target vpn gateway %s: %w", gw.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
