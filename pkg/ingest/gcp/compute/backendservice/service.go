package backendservice

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputebackendservice"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputebackendservicebackend"
)

// Service handles GCP Compute backend service ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new backend service ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for backend service ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of backend service ingestion.
type IngestResult struct {
	ProjectID           string
	BackendServiceCount int
	CollectedAt         time.Time
	DurationMillis      int64
}

// Ingest fetches backend services from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	// Fetch backend services from GCP
	backendServices, err := s.client.ListBackendServices(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list backend services: %w", err)
	}

	// Convert to data structs
	dataList := make([]*BackendServiceData, 0, len(backendServices))
	for _, bs := range backendServices {
		data, err := ConvertBackendService(bs, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert backend service: %w", err)
		}
		dataList = append(dataList, data)
	}

	// Save to database
	if err := s.saveBackendServices(ctx, dataList); err != nil {
		return nil, fmt.Errorf("failed to save backend services: %w", err)
	}

	return &IngestResult{
		ProjectID:           params.ProjectID,
		BackendServiceCount: len(dataList),
		CollectedAt:         collectedAt,
		DurationMillis:      time.Since(startTime).Milliseconds(),
	}, nil
}

// saveBackendServices saves backend services to the database with history tracking.
func (s *Service) saveBackendServices(ctx context.Context, backendServices []*BackendServiceData) error {
	if len(backendServices) == 0 {
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

	for _, data := range backendServices {
		// Load existing backend service with backends
		existing, err := tx.BronzeGCPComputeBackendService.Query().
			Where(bronzegcpcomputebackendservice.ID(data.ID)).
			WithBackends().
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing backend service %s: %w", data.ID, err)
		}

		// Compute diff
		diff := DiffBackendServiceData(existing, data)

		// Skip if no changes
		if !diff.HasAnyChange() && existing != nil {
			// Update collected_at only
			if err := tx.BronzeGCPComputeBackendService.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for backend service %s: %w", data.ID, err)
			}
			continue
		}

		// Delete old backends if updating
		if existing != nil {
			_, err := tx.BronzeGCPComputeBackendServiceBackend.Delete().
				Where(bronzegcpcomputebackendservicebackend.HasBackendServiceWith(bronzegcpcomputebackendservice.ID(data.ID))).
				Exec(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete old backends for backend service %s: %w", data.ID, err)
			}
		}

		// Create or update backend service
		var savedService *ent.BronzeGCPComputeBackendService
		if existing == nil {
			// Create new backend service
			create := tx.BronzeGCPComputeBackendService.Create().
				SetID(data.ID).
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
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			// Set JSON array fields if present
			if data.HealthChecksJSON != nil {
				create.SetHealthChecksJSON(data.HealthChecksJSON)
			}
			if data.LocalityLbPoliciesJSON != nil {
				create.SetLocalityLbPoliciesJSON(data.LocalityLbPoliciesJSON)
			}
			if data.UsedByJSON != nil {
				create.SetUsedByJSON(data.UsedByJSON)
			}
			if data.CustomRequestHeadersJSON != nil {
				create.SetCustomRequestHeadersJSON(data.CustomRequestHeadersJSON)
			}
			if data.CustomResponseHeadersJSON != nil {
				create.SetCustomResponseHeadersJSON(data.CustomResponseHeadersJSON)
			}

			// Set JSON object fields if present
			if data.CdnPolicyJSON != nil {
				create.SetCdnPolicyJSON(data.CdnPolicyJSON)
			}
			if data.CircuitBreakersJSON != nil {
				create.SetCircuitBreakersJSON(data.CircuitBreakersJSON)
			}
			if data.ConnectionDrainingJSON != nil {
				create.SetConnectionDrainingJSON(data.ConnectionDrainingJSON)
			}
			if data.ConnectionTrackingPolicyJSON != nil {
				create.SetConnectionTrackingPolicyJSON(data.ConnectionTrackingPolicyJSON)
			}
			if data.ConsistentHashJSON != nil {
				create.SetConsistentHashJSON(data.ConsistentHashJSON)
			}
			if data.FailoverPolicyJSON != nil {
				create.SetFailoverPolicyJSON(data.FailoverPolicyJSON)
			}
			if data.IapJSON != nil {
				create.SetIapJSON(data.IapJSON)
			}
			if data.LogConfigJSON != nil {
				create.SetLogConfigJSON(data.LogConfigJSON)
			}
			if data.MaxStreamDurationJSON != nil {
				create.SetMaxStreamDurationJSON(data.MaxStreamDurationJSON)
			}
			if data.OutlierDetectionJSON != nil {
				create.SetOutlierDetectionJSON(data.OutlierDetectionJSON)
			}
			if data.SecuritySettingsJSON != nil {
				create.SetSecuritySettingsJSON(data.SecuritySettingsJSON)
			}
			if data.SubsettingJSON != nil {
				create.SetSubsettingJSON(data.SubsettingJSON)
			}
			if data.ServiceBindingsJSON != nil {
				create.SetServiceBindingsJSON(data.ServiceBindingsJSON)
			}

			savedService, err = create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create backend service %s: %w", data.ID, err)
			}

			// Create backends for new backend service
			for _, backend := range data.Backends {
				_, err := tx.BronzeGCPComputeBackendServiceBackend.Create().
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
					SetBackendService(savedService).
					Save(ctx)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create backend for backend service %s: %w", data.ID, err)
				}
			}
		} else {
			// Update existing backend service
			update := tx.BronzeGCPComputeBackendService.UpdateOneID(data.ID).
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
				SetCollectedAt(data.CollectedAt)

			// Set JSON array fields if present
			if data.HealthChecksJSON != nil {
				update.SetHealthChecksJSON(data.HealthChecksJSON)
			}
			if data.LocalityLbPoliciesJSON != nil {
				update.SetLocalityLbPoliciesJSON(data.LocalityLbPoliciesJSON)
			}
			if data.UsedByJSON != nil {
				update.SetUsedByJSON(data.UsedByJSON)
			}
			if data.CustomRequestHeadersJSON != nil {
				update.SetCustomRequestHeadersJSON(data.CustomRequestHeadersJSON)
			}
			if data.CustomResponseHeadersJSON != nil {
				update.SetCustomResponseHeadersJSON(data.CustomResponseHeadersJSON)
			}

			// Set JSON object fields if present
			if data.CdnPolicyJSON != nil {
				update.SetCdnPolicyJSON(data.CdnPolicyJSON)
			}
			if data.CircuitBreakersJSON != nil {
				update.SetCircuitBreakersJSON(data.CircuitBreakersJSON)
			}
			if data.ConnectionDrainingJSON != nil {
				update.SetConnectionDrainingJSON(data.ConnectionDrainingJSON)
			}
			if data.ConnectionTrackingPolicyJSON != nil {
				update.SetConnectionTrackingPolicyJSON(data.ConnectionTrackingPolicyJSON)
			}
			if data.ConsistentHashJSON != nil {
				update.SetConsistentHashJSON(data.ConsistentHashJSON)
			}
			if data.FailoverPolicyJSON != nil {
				update.SetFailoverPolicyJSON(data.FailoverPolicyJSON)
			}
			if data.IapJSON != nil {
				update.SetIapJSON(data.IapJSON)
			}
			if data.LogConfigJSON != nil {
				update.SetLogConfigJSON(data.LogConfigJSON)
			}
			if data.MaxStreamDurationJSON != nil {
				update.SetMaxStreamDurationJSON(data.MaxStreamDurationJSON)
			}
			if data.OutlierDetectionJSON != nil {
				update.SetOutlierDetectionJSON(data.OutlierDetectionJSON)
			}
			if data.SecuritySettingsJSON != nil {
				update.SetSecuritySettingsJSON(data.SecuritySettingsJSON)
			}
			if data.SubsettingJSON != nil {
				update.SetSubsettingJSON(data.SubsettingJSON)
			}
			if data.ServiceBindingsJSON != nil {
				update.SetServiceBindingsJSON(data.ServiceBindingsJSON)
			}

			savedService, err = update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update backend service %s: %w", data.ID, err)
			}

			// Create new backends
			for _, backend := range data.Backends {
				_, err := tx.BronzeGCPComputeBackendServiceBackend.Create().
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
					SetBackendService(savedService).
					Save(ctx)
				if err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to create backend for backend service %s: %w", data.ID, err)
				}
			}
		}

		// Track history
		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for backend service %s: %w", data.ID, err)
			}
		} else {
			if err := s.history.UpdateHistory(ctx, tx, existing, data, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for backend service %s: %w", data.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleBackendServices removes backend services that were not collected in the latest run.
// Also closes history records for deleted backend services.
func (s *Service) DeleteStaleBackendServices(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	// Find stale backend services
	staleServices, err := tx.BronzeGCPComputeBackendService.Query().
		Where(
			bronzegcpcomputebackendservice.ProjectID(projectID),
			bronzegcpcomputebackendservice.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Close history and delete each stale backend service
	for _, svc := range staleServices {
		// Close history
		if err := s.history.CloseHistory(ctx, tx, svc.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for backend service %s: %w", svc.ID, err)
		}

		// Delete backend service (backends will be deleted automatically via CASCADE)
		if err := tx.BronzeGCPComputeBackendService.DeleteOne(svc).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete backend service %s: %w", svc.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
