package targetsslproxy

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzehistorygcpcomputetargetsslproxy"
)

// HistoryService manages target SSL proxy history tracking.
type HistoryService struct {
	entClient *ent.Client
}

// NewHistoryService creates a new history service.
func NewHistoryService(entClient *ent.Client) *HistoryService {
	return &HistoryService{entClient: entClient}
}

// CreateHistory creates a history record for a new target SSL proxy.
func (h *HistoryService) CreateHistory(ctx context.Context, tx *ent.Tx, data *TargetSslProxyData, now time.Time) error {
	create := tx.BronzeHistoryGCPComputeTargetSslProxy.Create().
		SetResourceID(data.ID).
		SetValidFrom(now).
		SetCollectedAt(data.CollectedAt).
		SetFirstCollectedAt(data.CollectedAt).
		SetName(data.Name).
		SetProjectID(data.ProjectID)

	if data.Description != "" {
		create.SetDescription(data.Description)
	}
	if data.CreationTimestamp != "" {
		create.SetCreationTimestamp(data.CreationTimestamp)
	}
	if data.SelfLink != "" {
		create.SetSelfLink(data.SelfLink)
	}
	if data.Service != "" {
		create.SetService(data.Service)
	}
	if data.ProxyHeader != "" {
		create.SetProxyHeader(data.ProxyHeader)
	}
	if data.CertificateMap != "" {
		create.SetCertificateMap(data.CertificateMap)
	}
	if data.SslPolicy != "" {
		create.SetSslPolicy(data.SslPolicy)
	}
	if data.SslCertificatesJSON != nil {
		create.SetSslCertificatesJSON(data.SslCertificatesJSON)
	}

	_, err := create.Save(ctx)
	return err
}

// UpdateHistory closes old history and creates new history based on diff.
func (h *HistoryService) UpdateHistory(ctx context.Context, tx *ent.Tx, old *ent.BronzeGCPComputeTargetSslProxy, new *TargetSslProxyData, diff *TargetSslProxyDiff, now time.Time) error {
	if !diff.IsChanged {
		return nil
	}

	_, err := tx.BronzeHistoryGCPComputeTargetSslProxy.Update().
		Where(
			bronzehistorygcpcomputetargetsslproxy.ResourceID(old.ID),
			bronzehistorygcpcomputetargetsslproxy.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("close old history: %w", err)
	}

	create := tx.BronzeHistoryGCPComputeTargetSslProxy.Create().
		SetResourceID(new.ID).
		SetValidFrom(now).
		SetCollectedAt(new.CollectedAt).
		SetFirstCollectedAt(old.FirstCollectedAt).
		SetName(new.Name).
		SetProjectID(new.ProjectID)

	if new.Description != "" {
		create.SetDescription(new.Description)
	}
	if new.CreationTimestamp != "" {
		create.SetCreationTimestamp(new.CreationTimestamp)
	}
	if new.SelfLink != "" {
		create.SetSelfLink(new.SelfLink)
	}
	if new.Service != "" {
		create.SetService(new.Service)
	}
	if new.ProxyHeader != "" {
		create.SetProxyHeader(new.ProxyHeader)
	}
	if new.CertificateMap != "" {
		create.SetCertificateMap(new.CertificateMap)
	}
	if new.SslPolicy != "" {
		create.SetSslPolicy(new.SslPolicy)
	}
	if new.SslCertificatesJSON != nil {
		create.SetSslCertificatesJSON(new.SslCertificatesJSON)
	}

	_, err = create.Save(ctx)
	return err
}

// CloseHistory closes history records for a deleted target SSL proxy.
func (h *HistoryService) CloseHistory(ctx context.Context, tx *ent.Tx, resourceID string, now time.Time) error {
	_, err := tx.BronzeHistoryGCPComputeTargetSslProxy.Update().
		Where(
			bronzehistorygcpcomputetargetsslproxy.ResourceID(resourceID),
			bronzehistorygcpcomputetargetsslproxy.ValidToIsNil(),
		).
		SetValidTo(now).
		Save(ctx)
	if ent.IsNotFound(err) {
		return nil
	}
	return err
}
