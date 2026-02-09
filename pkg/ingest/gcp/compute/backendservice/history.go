package backendservice

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputebackendservice"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputebackendservicebackend"
)

// HistoryService manages backend service history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates initial history records for a new backend service.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *BackendServiceData, now time.Time) error {
	// Create backend service history
	bsHistory, err := tx.BronzeHistoryGCPComputeBackendService.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetDescription(data.Description).
		SetCreationTimestamp(data.CreationTimestamp).
		SetSelfLink(data.SelfLink).
		SetFingerprint(data.Fingerprint).
		SetLoadBalancingScheme(data.LoadBalancingScheme).
		SetProtocol(data.Protocol).
		SetPortName(data.PortName).
		SetPort(data.Port).
		SetTimeoutSec(data.TimeoutSec).
		SetRegion(data.Region).
		SetNetwork(data.Network).
		SetSecurityPolicy(data.SecurityPolicy).
		SetEdgeSecurityPolicy(data.EdgeSecurityPolicy).
		SetSessionAffinity(data.SessionAffinity).
		SetAffinityCookieTTLSec(data.AffinityCookieTtlSec).
		SetLocalityLbPolicy(data.LocalityLbPolicy).
		SetCompressionMode(data.CompressionMode).
		SetServiceLbPolicy(data.ServiceLbPolicy).
		SetEnableCdn(data.EnableCdn).
		SetProjectID(data.ProjectID).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to create backend service history: %w", err)
	}

	// Set JSON array fields if present
	if data.HealthChecksJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetHealthChecksJSON(data.HealthChecksJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set health checks json: %w", err)
		}
	}
	if data.LocalityLbPoliciesJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetLocalityLbPoliciesJSON(data.LocalityLbPoliciesJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set locality lb policies json: %w", err)
		}
	}
	if data.UsedByJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetUsedByJSON(data.UsedByJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set used by json: %w", err)
		}
	}
	if data.CustomRequestHeadersJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetCustomRequestHeadersJSON(data.CustomRequestHeadersJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set custom request headers json: %w", err)
		}
	}
	if data.CustomResponseHeadersJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetCustomResponseHeadersJSON(data.CustomResponseHeadersJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set custom response headers json: %w", err)
		}
	}

	// Set JSON object fields if present
	if data.CdnPolicyJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetCdnPolicyJSON(data.CdnPolicyJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set cdn policy json: %w", err)
		}
	}
	if data.CircuitBreakersJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetCircuitBreakersJSON(data.CircuitBreakersJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set circuit breakers json: %w", err)
		}
	}
	if data.ConnectionDrainingJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetConnectionDrainingJSON(data.ConnectionDrainingJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set connection draining json: %w", err)
		}
	}
	if data.ConnectionTrackingPolicyJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetConnectionTrackingPolicyJSON(data.ConnectionTrackingPolicyJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set connection tracking policy json: %w", err)
		}
	}
	if data.ConsistentHashJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetConsistentHashJSON(data.ConsistentHashJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set consistent hash json: %w", err)
		}
	}
	if data.FailoverPolicyJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetFailoverPolicyJSON(data.FailoverPolicyJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set failover policy json: %w", err)
		}
	}
	if data.IapJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetIapJSON(data.IapJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set iap json: %w", err)
		}
	}
	if data.LogConfigJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetLogConfigJSON(data.LogConfigJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set log config json: %w", err)
		}
	}
	if data.MaxStreamDurationJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetMaxStreamDurationJSON(data.MaxStreamDurationJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set max stream duration json: %w", err)
		}
	}
	if data.OutlierDetectionJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetOutlierDetectionJSON(data.OutlierDetectionJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set outlier detection json: %w", err)
		}
	}
	if data.SecuritySettingsJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetSecuritySettingsJSON(data.SecuritySettingsJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set security settings json: %w", err)
		}
	}
	if data.SubsettingJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetSubsettingJSON(data.SubsettingJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set subsetting json: %w", err)
		}
	}
	if data.ServiceBindingsJSON != nil {
		if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(bsHistory).
			SetServiceBindingsJSON(data.ServiceBindingsJSON).
			Save(ctx); err != nil {
			return fmt.Errorf("failed to set service bindings json: %w", err)
		}
	}

	// Create backend history records
	for _, backend := range data.Backends {
		_, err := tx.BronzeHistoryGCPComputeBackendServiceBackend.Create().
			SetBackendServiceHistoryID(bsHistory.HistoryID).
			SetValidFrom(now).
			SetGroup(backend.Group).
			SetBalancingMode(backend.BalancingMode).
			SetCapacityScaler(backend.CapacityScaler).
			SetDescription(backend.Description).
			SetFailover(backend.Failover).
			SetMaxConnections(backend.MaxConnections).
			SetMaxConnectionsPerEndpoint(backend.MaxConnectionsPerEndpoint).
			SetMaxConnectionsPerInstance(backend.MaxConnectionsPerInstance).
			SetMaxRate(backend.MaxRate).
			SetMaxRatePerEndpoint(backend.MaxRatePerEndpoint).
			SetMaxRatePerInstance(backend.MaxRatePerInstance).
			SetMaxUtilization(backend.MaxUtilization).
			SetPreference(backend.Preference).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create backend history: %w", err)
		}
	}

	return nil
}

// UpdateHistory updates history records for a changed backend service.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeBackendService, new *BackendServiceData, diff *BackendServiceDiff, now time.Time) error {
	// Get current backend service history
	currentHistory, err := tx.BronzeHistoryGCPComputeBackendService.Query().
		Where(
			bronzehistorygcpcomputebackendservice.ResourceID(old.ID),
			bronzehistorygcpcomputebackendservice.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		return fmt.Errorf("failed to find current backend service history: %w", err)
	}

	// Close current backend service history if core fields changed
	if diff.IsChanged {
		// Close old backend history first
		_, err := tx.BronzeHistoryGCPComputeBackendServiceBackend.Update().
			Where(
				bronzehistorygcpcomputebackendservicebackend.BackendServiceHistoryID(currentHistory.HistoryID),
				bronzehistorygcpcomputebackendservicebackend.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close old backend history: %w", err)
		}

		// Close current backend service history
		err = tx.BronzeHistoryGCPComputeBackendService.UpdateOne(currentHistory).
			SetValidTo(now).
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("failed to close current backend service history: %w", err)
		}

		// Create new backend service history
		newHistory, err := tx.BronzeHistoryGCPComputeBackendService.Create().
			SetResourceID(new.ID).
			SetValidFrom(now).
			SetCollectedAt(new.CollectedAt).
			SetFirstCollectedAt(old.FirstCollectedAt).
			SetName(new.Name).
			SetDescription(new.Description).
			SetCreationTimestamp(new.CreationTimestamp).
			SetSelfLink(new.SelfLink).
			SetFingerprint(new.Fingerprint).
			SetLoadBalancingScheme(new.LoadBalancingScheme).
			SetProtocol(new.Protocol).
			SetPortName(new.PortName).
			SetPort(new.Port).
			SetTimeoutSec(new.TimeoutSec).
			SetRegion(new.Region).
			SetNetwork(new.Network).
			SetSecurityPolicy(new.SecurityPolicy).
			SetEdgeSecurityPolicy(new.EdgeSecurityPolicy).
			SetSessionAffinity(new.SessionAffinity).
			SetAffinityCookieTTLSec(new.AffinityCookieTtlSec).
			SetLocalityLbPolicy(new.LocalityLbPolicy).
			SetCompressionMode(new.CompressionMode).
			SetServiceLbPolicy(new.ServiceLbPolicy).
			SetEnableCdn(new.EnableCdn).
			SetProjectID(new.ProjectID).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to create new backend service history: %w", err)
		}

		// Set JSON array fields if present
		if new.HealthChecksJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetHealthChecksJSON(new.HealthChecksJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set health checks json: %w", err)
			}
		}
		if new.LocalityLbPoliciesJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetLocalityLbPoliciesJSON(new.LocalityLbPoliciesJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set locality lb policies json: %w", err)
			}
		}
		if new.UsedByJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetUsedByJSON(new.UsedByJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set used by json: %w", err)
			}
		}
		if new.CustomRequestHeadersJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetCustomRequestHeadersJSON(new.CustomRequestHeadersJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set custom request headers json: %w", err)
			}
		}
		if new.CustomResponseHeadersJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetCustomResponseHeadersJSON(new.CustomResponseHeadersJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set custom response headers json: %w", err)
			}
		}

		// Set JSON object fields if present
		if new.CdnPolicyJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetCdnPolicyJSON(new.CdnPolicyJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set cdn policy json: %w", err)
			}
		}
		if new.CircuitBreakersJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetCircuitBreakersJSON(new.CircuitBreakersJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set circuit breakers json: %w", err)
			}
		}
		if new.ConnectionDrainingJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetConnectionDrainingJSON(new.ConnectionDrainingJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set connection draining json: %w", err)
			}
		}
		if new.ConnectionTrackingPolicyJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetConnectionTrackingPolicyJSON(new.ConnectionTrackingPolicyJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set connection tracking policy json: %w", err)
			}
		}
		if new.ConsistentHashJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetConsistentHashJSON(new.ConsistentHashJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set consistent hash json: %w", err)
			}
		}
		if new.FailoverPolicyJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetFailoverPolicyJSON(new.FailoverPolicyJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set failover policy json: %w", err)
			}
		}
		if new.IapJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetIapJSON(new.IapJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set iap json: %w", err)
			}
		}
		if new.LogConfigJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetLogConfigJSON(new.LogConfigJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set log config json: %w", err)
			}
		}
		if new.MaxStreamDurationJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetMaxStreamDurationJSON(new.MaxStreamDurationJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set max stream duration json: %w", err)
			}
		}
		if new.OutlierDetectionJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetOutlierDetectionJSON(new.OutlierDetectionJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set outlier detection json: %w", err)
			}
		}
		if new.SecuritySettingsJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetSecuritySettingsJSON(new.SecuritySettingsJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set security settings json: %w", err)
			}
		}
		if new.SubsettingJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetSubsettingJSON(new.SubsettingJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set subsetting json: %w", err)
			}
		}
		if new.ServiceBindingsJSON != nil {
			if _, err := tx.BronzeHistoryGCPComputeBackendService.UpdateOne(newHistory).
				SetServiceBindingsJSON(new.ServiceBindingsJSON).
				Save(ctx); err != nil {
				return fmt.Errorf("failed to set service bindings json: %w", err)
			}
		}

		// Create new backend history linked to new backend service history
		for _, backend := range new.Backends {
			_, err := tx.BronzeHistoryGCPComputeBackendServiceBackend.Create().
				SetBackendServiceHistoryID(newHistory.HistoryID).
				SetValidFrom(now).
				SetGroup(backend.Group).
				SetBalancingMode(backend.BalancingMode).
				SetCapacityScaler(backend.CapacityScaler).
				SetDescription(backend.Description).
				SetFailover(backend.Failover).
				SetMaxConnections(backend.MaxConnections).
				SetMaxConnectionsPerEndpoint(backend.MaxConnectionsPerEndpoint).
				SetMaxConnectionsPerInstance(backend.MaxConnectionsPerInstance).
				SetMaxRate(backend.MaxRate).
				SetMaxRatePerEndpoint(backend.MaxRatePerEndpoint).
				SetMaxRatePerInstance(backend.MaxRatePerInstance).
				SetMaxUtilization(backend.MaxUtilization).
				SetPreference(backend.Preference).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create backend history: %w", err)
			}
		}
	} else if diff.BackendsDiff.HasChanges {
		// Only backends changed - close old backend history and create new ones
		_, err := tx.BronzeHistoryGCPComputeBackendServiceBackend.Update().
			Where(
				bronzehistorygcpcomputebackendservicebackend.BackendServiceHistoryID(currentHistory.HistoryID),
				bronzehistorygcpcomputebackendservicebackend.ValidToIsNil(),
			).
			SetValidTo(now).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("failed to close backend history: %w", err)
		}

		for _, backend := range new.Backends {
			_, err := tx.BronzeHistoryGCPComputeBackendServiceBackend.Create().
				SetBackendServiceHistoryID(currentHistory.HistoryID).
				SetValidFrom(now).
				SetGroup(backend.Group).
				SetBalancingMode(backend.BalancingMode).
				SetCapacityScaler(backend.CapacityScaler).
				SetDescription(backend.Description).
				SetFailover(backend.Failover).
				SetMaxConnections(backend.MaxConnections).
				SetMaxConnectionsPerEndpoint(backend.MaxConnectionsPerEndpoint).
				SetMaxConnectionsPerInstance(backend.MaxConnectionsPerInstance).
				SetMaxRate(backend.MaxRate).
				SetMaxRatePerEndpoint(backend.MaxRatePerEndpoint).
				SetMaxRatePerInstance(backend.MaxRatePerInstance).
				SetMaxUtilization(backend.MaxUtilization).
				SetPreference(backend.Preference).
				Save(ctx)
			if err != nil {
				return fmt.Errorf("failed to create backend history: %w", err)
			}
		}
	}

	return nil
}

// CloseHistory closes all history records for a deleted backend service.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	// Get current backend service history
	currentHistory, err := tx.BronzeHistoryGCPComputeBackendService.Query().
		Where(
			bronzehistorygcpcomputebackendservice.ResourceID(resourceID),
			bronzehistorygcpcomputebackendservice.ValidToIsNil(),
		).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil // No history to close
		}
		return fmt.Errorf("failed to find current backend service history: %w", err)
	}

	// Close backend service history
	err = tx.BronzeHistoryGCPComputeBackendService.UpdateOne(currentHistory).
		SetValidTo(now).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to close backend service history: %w", err)
	}

	// Close backend history
	_, err = tx.BronzeHistoryGCPComputeBackendServiceBackend.Update().
		Where(
			bronzehistorygcpcomputebackendservicebackend.BackendServiceHistoryID(currentHistory.HistoryID),
			bronzehistorygcpcomputebackendservicebackend.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("failed to close backend history: %w", err)
	}

	return nil
}
