package negendpoint

import (
	"context"
	"fmt"
	"time"

	"hotpot/pkg/storage/ent"
	"hotpot/pkg/storage/ent/bronzegcpcomputenegendpoint"
)

type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

type IngestParams struct {
	ProjectID string
}

type IngestResult struct {
	ProjectID        string
	NegEndpointCount int
	CollectedAt      time.Time
	DurationMillis   int64
}

func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	endpoints, err := s.client.ListNegEndpoints(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list NEG endpoints: %w", err)
	}

	epDataList := make([]*NegEndpointData, 0, len(endpoints))
	for _, ewn := range endpoints {
		data, err := ConvertNegEndpoint(ewn, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert NEG endpoint: %w", err)
		}
		epDataList = append(epDataList, data)
	}

	if err := s.saveNegEndpoints(ctx, epDataList); err != nil {
		return nil, fmt.Errorf("failed to save NEG endpoints: %w", err)
	}

	return &IngestResult{
		ProjectID:        params.ProjectID,
		NegEndpointCount: len(epDataList),
		CollectedAt:      collectedAt,
		DurationMillis:   time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveNegEndpoints(ctx context.Context, endpoints []*NegEndpointData) error {
	if len(endpoints) == 0 {
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

	for _, data := range endpoints {
		existing, err := tx.BronzeGCPComputeNegEndpoint.Query().
			Where(bronzegcpcomputenegendpoint.ID(data.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing NEG endpoint %s: %w", data.ID, err)
		}

		diff := DiffNegEndpointData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPComputeNegEndpoint.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for NEG endpoint %s: %w", data.ID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGCPComputeNegEndpoint.Create().
				SetID(data.ID).
				SetInstance(data.Instance).
				SetIPAddress(data.IpAddress).
				SetIpv6Address(data.Ipv6Address).
				SetPort(data.Port).
				SetFqdn(data.Fqdn).
				SetNegName(data.NegName).
				SetZone(data.Zone).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.AnnotationsJSON != nil {
				create.SetAnnotationsJSON(data.AnnotationsJSON)
			}

			if _, err := create.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create NEG endpoint %s: %w", data.ID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for NEG endpoint %s: %w", data.ID, err)
			}
		} else {
			update := tx.BronzeGCPComputeNegEndpoint.UpdateOneID(data.ID).
				SetInstance(data.Instance).
				SetIPAddress(data.IpAddress).
				SetIpv6Address(data.Ipv6Address).
				SetPort(data.Port).
				SetFqdn(data.Fqdn).
				SetNegName(data.NegName).
				SetZone(data.Zone).
				SetCollectedAt(data.CollectedAt)

			if data.AnnotationsJSON != nil {
				update.SetAnnotationsJSON(data.AnnotationsJSON)
			}

			if _, err := update.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update NEG endpoint %s: %w", data.ID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for NEG endpoint %s: %w", data.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Service) DeleteStaleNegEndpoints(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleEndpoints, err := tx.BronzeGCPComputeNegEndpoint.Query().
		Where(
			bronzegcpcomputenegendpoint.ProjectID(projectID),
			bronzegcpcomputenegendpoint.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, ep := range staleEndpoints {
		if err := s.history.CloseHistory(ctx, tx, ep.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for NEG endpoint %s: %w", ep.ID, err)
		}

		if err := tx.BronzeGCPComputeNegEndpoint.DeleteOne(ep).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete NEG endpoint %s: %w", ep.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
