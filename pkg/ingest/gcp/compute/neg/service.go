package neg

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcpcomputeneg"
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
	ProjectID      string
	NegCount       int
	CollectedAt    time.Time
	DurationMillis int64
}

func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	negs, err := s.client.ListNegs(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list NEGs: %w", err)
	}

	negDataList := make([]*NegData, 0, len(negs))
	for _, neg := range negs {
		data, err := ConvertNeg(neg, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert NEG: %w", err)
		}
		negDataList = append(negDataList, data)
	}

	if err := s.saveNegs(ctx, negDataList); err != nil {
		return nil, fmt.Errorf("failed to save NEGs: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		NegCount:       len(negDataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveNegs(ctx context.Context, negs []*NegData) error {
	if len(negs) == 0 {
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

	for _, data := range negs {
		existing, err := tx.BronzeGCPComputeNeg.Query().
			Where(bronzegcpcomputeneg.ID(data.ID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing NEG %s: %w", data.ID, err)
		}

		diff := DiffNegData(existing, data)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPComputeNeg.UpdateOneID(data.ID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for NEG %s: %w", data.ID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeGCPComputeNeg.Create().
				SetID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetCreationTimestamp(data.CreationTimestamp).
				SetSelfLink(data.SelfLink).
				SetNetwork(data.Network).
				SetSubnetwork(data.Subnetwork).
				SetZone(data.Zone).
				SetNetworkEndpointType(data.NetworkEndpointType).
				SetDefaultPort(data.DefaultPort).
				SetSize(data.Size).
				SetRegion(data.Region).
				SetProjectID(data.ProjectID).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.AnnotationsJSON != nil {
				create.SetAnnotationsJSON(data.AnnotationsJSON)
			}
			if data.AppEngineJSON != nil {
				create.SetAppEngineJSON(data.AppEngineJSON)
			}
			if data.CloudFunctionJSON != nil {
				create.SetCloudFunctionJSON(data.CloudFunctionJSON)
			}
			if data.CloudRunJSON != nil {
				create.SetCloudRunJSON(data.CloudRunJSON)
			}
			if data.PscDataJSON != nil {
				create.SetPscDataJSON(data.PscDataJSON)
			}

			if _, err := create.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create NEG %s: %w", data.ID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for NEG %s: %w", data.ID, err)
			}
		} else {
			update := tx.BronzeGCPComputeNeg.UpdateOneID(data.ID).
				SetName(data.Name).
				SetDescription(data.Description).
				SetCreationTimestamp(data.CreationTimestamp).
				SetSelfLink(data.SelfLink).
				SetNetwork(data.Network).
				SetSubnetwork(data.Subnetwork).
				SetZone(data.Zone).
				SetNetworkEndpointType(data.NetworkEndpointType).
				SetDefaultPort(data.DefaultPort).
				SetSize(data.Size).
				SetRegion(data.Region).
				SetCollectedAt(data.CollectedAt)

			if data.AnnotationsJSON != nil {
				update.SetAnnotationsJSON(data.AnnotationsJSON)
			}
			if data.AppEngineJSON != nil {
				update.SetAppEngineJSON(data.AppEngineJSON)
			}
			if data.CloudFunctionJSON != nil {
				update.SetCloudFunctionJSON(data.CloudFunctionJSON)
			}
			if data.CloudRunJSON != nil {
				update.SetCloudRunJSON(data.CloudRunJSON)
			}
			if data.PscDataJSON != nil {
				update.SetPscDataJSON(data.PscDataJSON)
			}

			if _, err := update.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update NEG %s: %w", data.ID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for NEG %s: %w", data.ID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Service) DeleteStaleNegs(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	staleNegs, err := tx.BronzeGCPComputeNeg.Query().
		Where(
			bronzegcpcomputeneg.ProjectID(projectID),
			bronzegcpcomputeneg.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, neg := range staleNegs {
		if err := s.history.CloseHistory(ctx, tx, neg.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for NEG %s: %w", neg.ID, err)
		}

		if err := tx.BronzeGCPComputeNeg.DeleteOne(neg).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete NEG %s: %w", neg.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
