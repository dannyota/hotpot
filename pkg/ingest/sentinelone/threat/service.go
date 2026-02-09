package threat

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzes1threat"
)

// Service handles SentinelOne threat ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new threat ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestResult contains the result of threat ingestion.
type IngestResult struct {
	ThreatCount    int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches all threats from SentinelOne using cursor pagination.
func (s *Service) Ingest(ctx context.Context, heartbeat func()) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	var allThreats []*ThreatData
	cursor := ""

	for {
		batch, err := s.client.GetThreatsBatch(cursor)
		if err != nil {
			return nil, fmt.Errorf("get threats batch: %w", err)
		}

		for _, apiThreat := range batch.Threats {
			allThreats = append(allThreats, ConvertThreat(apiThreat, collectedAt))
		}

		if heartbeat != nil {
			heartbeat()
		}

		if !batch.HasMore {
			break
		}
		cursor = batch.NextCursor
	}

	if err := s.saveThreats(ctx, allThreats); err != nil {
		return nil, fmt.Errorf("save threats: %w", err)
	}

	return &IngestResult{
		ThreatCount:    len(allThreats),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

func (s *Service) saveThreats(ctx context.Context, threats []*ThreatData) error {
	if len(threats) == 0 {
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

	for _, data := range threats {
		existing, err := tx.BronzeS1Threat.Query().
			Where(bronzes1threat.ID(data.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("load existing threat %s: %w", data.ResourceID, err)
		}

		diff := DiffThreatData(existing, data)

		if !diff.IsNew && !diff.IsChanged {
			if err := tx.BronzeS1Threat.UpdateOneID(data.ResourceID).
				SetCollectedAt(data.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update collected_at for threat %s: %w", data.ResourceID, err)
			}
			continue
		}

		if existing == nil {
			create := tx.BronzeS1Threat.Create().
				SetID(data.ResourceID).
				SetAgentID(data.AgentID).
				SetClassification(data.Classification).
				SetThreatName(data.ThreatName).
				SetFilePath(data.FilePath).
				SetStatus(data.Status).
				SetAnalystVerdict(data.AnalystVerdict).
				SetConfidenceLevel(data.ConfidenceLevel).
				SetInitiatedBy(data.InitiatedBy).
				SetCollectedAt(data.CollectedAt).
				SetFirstCollectedAt(data.CollectedAt)

			if data.APICreatedAt != nil {
				create.SetAPICreatedAt(*data.APICreatedAt)
			}
			if data.ThreatInfoJSON != nil {
				create.SetThreatInfoJSON(data.ThreatInfoJSON)
			}

			if _, err := create.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("create threat %s: %w", data.ResourceID, err)
			}

			if err := s.history.CreateHistory(ctx, tx, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("create history for threat %s: %w", data.ResourceID, err)
			}
		} else {
			update := tx.BronzeS1Threat.UpdateOneID(data.ResourceID).
				SetAgentID(data.AgentID).
				SetClassification(data.Classification).
				SetThreatName(data.ThreatName).
				SetFilePath(data.FilePath).
				SetStatus(data.Status).
				SetAnalystVerdict(data.AnalystVerdict).
				SetConfidenceLevel(data.ConfidenceLevel).
				SetInitiatedBy(data.InitiatedBy).
				SetCollectedAt(data.CollectedAt)

			if data.APICreatedAt != nil {
				update.SetAPICreatedAt(*data.APICreatedAt)
			}
			if data.ThreatInfoJSON != nil {
				update.SetThreatInfoJSON(data.ThreatInfoJSON)
			}

			if _, err := update.Save(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("update threat %s: %w", data.ResourceID, err)
			}

			if err := s.history.UpdateHistory(ctx, tx, existing, data, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("update history for threat %s: %w", data.ResourceID, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// DeleteStale removes threats that were not collected in the latest run.
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

	stale, err := tx.BronzeS1Threat.Query().
		Where(bronzes1threat.CollectedAtLT(collectedAt)).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, t := range stale {
		if err := s.history.CloseHistory(ctx, tx, t.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("close history for threat %s: %w", t.ID, err)
		}

		if err := tx.BronzeS1Threat.DeleteOne(t).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("delete threat %s: %w", t.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
