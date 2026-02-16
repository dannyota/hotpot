package loadbalancer

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorydoloadbalancer"
)

// HistoryService handles history tracking for Load Balancers.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

func (h *HistoryService) buildCreate(tx *ent.Tx, data *LoadBalancerData) *ent.BronzeHistoryDOLoadBalancerCreate {
	create := tx.BronzeHistoryDOLoadBalancer.Create().
		SetResourceID(data.ResourceID).
		SetName(data.Name).
		SetIP(data.IP).
		SetIpv6(data.Ipv6).
		SetSizeSlug(data.SizeSlug).
		SetSizeUnit(data.SizeUnit).
		SetLbType(data.LbType).
		SetAlgorithm(data.Algorithm).
		SetStatus(data.Status).
		SetRegion(data.Region).
		SetTag(data.Tag).
		SetRedirectHTTPToHTTPS(data.RedirectHTTPToHTTPS).
		SetEnableProxyProtocol(data.EnableProxyProtocol).
		SetEnableBackendKeepalive(data.EnableBackendKeepalive).
		SetVpcUUID(data.VpcUUID).
		SetProjectID(data.ProjectID).
		SetNillableHTTPIdleTimeoutSeconds(data.HTTPIdleTimeoutSeconds).
		SetNillableDisableLetsEncryptDNSRecords(data.DisableLetsEncryptDNSRecords).
		SetNetwork(data.Network).
		SetNetworkStack(data.NetworkStack).
		SetTLSCipherPolicy(data.TLSCipherPolicy).
		SetAPICreatedAt(data.APICreatedAt).
		SetForwardingRulesJSON(data.ForwardingRulesJSON).
		SetHealthCheckJSON(data.HealthCheckJSON).
		SetStickySessionsJSON(data.StickySessionsJSON).
		SetFirewallJSON(data.FirewallJSON).
		SetDomainsJSON(data.DomainsJSON).
		SetGlbSettingsJSON(data.GlbSettingsJSON).
		SetDropletIdsJSON(data.DropletIdsJSON).
		SetTagsJSON(data.TagsJSON).
		SetTargetLoadBalancerIdsJSON(data.TargetLoadBalancerIdsJSON)

	return create
}

// CreateHistory creates a history record for a new Load Balancer.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *LoadBalancerData, now time.Time) error {
	_, err := h.buildCreate(tx, data).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create load balancer history: %w", err)
	}
	return nil
}

// UpdateHistory closes old history and creates new for a changed Load Balancer.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeDOLoadBalancer, new *LoadBalancerData, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOLoadBalancer.Query().
		Where(
			bronzehistorydoloadbalancer.ResourceID(old.ID),
			bronzehistorydoloadbalancer.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("find current load balancer history: %w", err)
	}

	if err := tx.BronzeHistoryDOLoadBalancer.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close load balancer history: %w", err)
	}

	_, err = h.buildCreate(tx, new).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("create new load balancer history: %w", err)
	}

	return nil
}

// CloseHistory closes history records for a deleted Load Balancer.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	currentHist, err := tx.BronzeHistoryDOLoadBalancer.Query().
		Where(
			bronzehistorydoloadbalancer.ResourceID(resourceID),
			bronzehistorydoloadbalancer.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("find current load balancer history: %w", err)
	}

	if err := tx.BronzeHistoryDOLoadBalancer.UpdateOne(currentHist).
		SetValidTo(now).
		Exec(ctx); err != nil {
		return fmt.Errorf("close load balancer history: %w", err)
	}

	return nil
}
