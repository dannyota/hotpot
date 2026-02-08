package negendpoint

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzehistorygcpcomputenegendpoint"
)

type HistoryService struct {
	entClient *ent.Client
}

func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *NegEndpointData, now time.Time) error {
	create := tx.BronzeHistoryGCPComputeNegEndpoint.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetInstance(data.Instance).
		SetIPAddress(data.IpAddress).
		SetIpv6Address(data.Ipv6Address).
		SetPort(data.Port).
		SetFqdn(data.Fqdn).
		SetNegName(data.NegName).
		SetZone(data.Zone).
		SetProjectID(data.ProjectID)

	if data.AnnotationsJSON != nil {
		create.SetAnnotationsJSON(data.AnnotationsJSON)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("failed to create NEG endpoint history: %w", err)
	}
	return nil
}

func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeNegEndpoint, new *NegEndpointData, now time.Time) error {
	// Close old history
	_, err := tx.BronzeHistoryGCPComputeNegEndpoint.Update().
		Where(
			bronzehistorygcpcomputenegendpoint.ResourceID(old.ID),
			bronzehistorygcpcomputenegendpoint.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close NEG endpoint history: %w", err)
	}

	// Create new history
	create := tx.BronzeHistoryGCPComputeNegEndpoint.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetInstance(new.Instance).
		SetIPAddress(new.IpAddress).
		SetIpv6Address(new.Ipv6Address).
		SetPort(new.Port).
		SetFqdn(new.Fqdn).
		SetNegName(new.NegName).
		SetZone(new.Zone).
		SetProjectID(new.ProjectID)

	if new.AnnotationsJSON != nil {
		create.SetAnnotationsJSON(new.AnnotationsJSON)
	}

	if _, err := create.Save(ctx); err != nil {
		return fmt.Errorf("failed to create new NEG endpoint history: %w", err)
	}

	return nil
}

func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeNegEndpoint.Update().
		Where(
			bronzehistorygcpcomputenegendpoint.ResourceID(resourceID),
			bronzehistorygcpcomputenegendpoint.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil
	}
	return err
}
