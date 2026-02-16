package loadbalancer

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzedoloadbalancer"
)

// Service handles DigitalOcean Load Balancer ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new Load Balancer ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of Load Balancer ingestion.
type IngestResult struct {
	LoadBalancerCount int
	CollectedAt       time.Time
	DurationMillis    int64
}

// Ingest fetches all Load Balancers from DigitalOcean and saves them.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	apiLBs, err := s.client.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list load balancers: %w", err)
	}

	if heartbeat != nil {
		heartbeat()
	}

	var allLBs []*LoadBalancerData
	for _, v := range apiLBs {
		allLBs = append(allLBs, ConvertLoadBalancer(v, collectedAt))
	}

	if err := s.saveLoadBalancers(ctx, allLBs); err != nil {
		return nil, fmt.Errorf("save load balancers: %w", err)
	}

	return &IngestResult{
		LoadBalancerCount: len(allLBs),
		CollectedAt:       collectedAt,
		DurationMillis:    time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveLoadBalancers(ctx context.Context, lbs []*LoadBalancerData) error {
	if len(lbs) == 0 {
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

	for _, data := range lbs {
		existing, err := tx.BronzeDOLoadBalancer.Query().
			Where(bronzedoloadbalancer.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing load balancer %s: %w", data.ResourceID, err)
		}

		diff := DiffLoadBalancerData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeDOLoadBalancer.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for load balancer %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			if _, err := tx.BronzeDOLoadBalancer.Create().
				SetID(data.ResourceID).
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
				SetTargetLoadBalancerIdsJSON(data.TargetLoadBalancerIdsJSON).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create load balancer %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for load balancer %s: %w", data.ResourceID, err)
			}
		} else {
			if _, err := tx.BronzeDOLoadBalancer.UpdateOneID(data.ResourceID).
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
				SetTargetLoadBalancerIdsJSON(data.TargetLoadBalancerIdsJSON).
				SetCollectedAt(data.CollectedAt).
				Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update load balancer %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for load balancer %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes Load Balancers that were not collected in the latest run.
func (s *Service) DeleteStale(ctx context.Context, collectedAt time.Time) error {
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

	stale, err := tx.BronzeDOLoadBalancer.Query().
		Where(bronzedoloadbalancer.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, doLB := range stale {
		if err := s.history.CloseHistory(ctx, tx, doLB.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for load balancer %s: %w", doLB.ID, err)
		}

		if err := tx.BronzeDOLoadBalancer.DeleteOne(doLB).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete load balancer %s: %w", doLB.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
