package logmetric

import (
	"context"
	"fmt"
	"time"

	"github.com/dannyota/hotpot/pkg/storage/ent"
	"github.com/dannyota/hotpot/pkg/storage/ent/bronzegcplogginglogmetric"
)

// Service handles GCP Cloud Logging log metric ingestion.
type Service struct {
	client    *Client
	entClient *ent.Client
	history   *HistoryService
}

// NewService creates a new log metric ingestion service.
func NewService(client *Client, entClient *ent.Client) *Service {
	return &Service{
		client:    client,
		entClient: entClient,
		history:   NewHistoryService(entClient),
	}
}

// IngestParams contains parameters for log metric ingestion.
type IngestParams struct {
	ProjectID string
}

// IngestResult contains the result of log metric ingestion.
type IngestResult struct {
	ProjectID      string
	LogMetricCount int
	CollectedAt    time.Time
	DurationMillis int64
}

// Ingest fetches log metrics from GCP and stores them in the bronze layer.
func (s *Service) Ingest(ctx context.Context, params IngestParams) (*IngestResult, error) {
	startTime := time.Now()
	collectedAt := startTime

	metrics, err := s.client.ListLogMetrics(ctx, params.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list log metrics: %w", err)
	}

	dataList := make([]*LogMetricData, 0, len(metrics))
	for _, m := range metrics {
		data, err := ConvertLogMetric(m, params.ProjectID, collectedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to convert log metric: %w", err)
		}
		dataList = append(dataList, data)
	}

	if err := s.saveLogMetrics(ctx, dataList); err != nil {
		return nil, fmt.Errorf("failed to save log metrics: %w", err)
	}

	return &IngestResult{
		ProjectID:      params.ProjectID,
		LogMetricCount: len(dataList),
		CollectedAt:    collectedAt,
		DurationMillis: time.Since(startTime).Milliseconds(),
	}, nil
}

// saveLogMetrics saves log metrics to the database with history tracking.
func (s *Service) saveLogMetrics(ctx context.Context, metrics []*LogMetricData) error {
	if len(metrics) == 0 {
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

	for _, metricData := range metrics {
		existing, err := tx.BronzeGCPLoggingLogMetric.Query().
			Where(bronzegcplogginglogmetric.ID(metricData.ResourceID)).
			First(ctx)
		if err != nil && !ent.IsNotFound(err) {
			tx.Rollback()
			return fmt.Errorf("failed to load existing log metric %s: %w", metricData.Name, err)
		}

		diff := DiffLogMetricData(existing, metricData)

		if !diff.HasAnyChange() && existing != nil {
			if err := tx.BronzeGCPLoggingLogMetric.UpdateOneID(metricData.ResourceID).
				SetCollectedAt(metricData.CollectedAt).
				Exec(ctx); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update collected_at for log metric %s: %w", metricData.Name, err)
			}
			continue
		}

		if existing == nil {
			// Create new log metric
			create := tx.BronzeGCPLoggingLogMetric.Create().
				SetID(metricData.ResourceID).
				SetName(metricData.Name).
				SetDescription(metricData.Description).
				SetFilter(metricData.Filter).
				SetValueExtractor(metricData.ValueExtractor).
				SetVersion(metricData.Version).
				SetDisabled(metricData.Disabled).
				SetCreateTime(metricData.CreateTime).
				SetUpdateTime(metricData.UpdateTime).
				SetProjectID(metricData.ProjectID).
				SetCollectedAt(metricData.CollectedAt).
				SetFirstCollectedAt(metricData.CollectedAt)

			if metricData.MetricDescriptorJSON != nil {
				create.SetMetricDescriptorJSON(metricData.MetricDescriptorJSON)
			}
			if metricData.LabelExtractorsJSON != nil {
				create.SetLabelExtractorsJSON(metricData.LabelExtractorsJSON)
			}
			if metricData.BucketOptionsJSON != nil {
				create.SetBucketOptionsJSON(metricData.BucketOptionsJSON)
			}

			_, err := create.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create log metric %s: %w", metricData.Name, err)
			}
		} else {
			// Update existing log metric
			update := tx.BronzeGCPLoggingLogMetric.UpdateOneID(metricData.ResourceID).
				SetName(metricData.Name).
				SetDescription(metricData.Description).
				SetFilter(metricData.Filter).
				SetValueExtractor(metricData.ValueExtractor).
				SetVersion(metricData.Version).
				SetDisabled(metricData.Disabled).
				SetCreateTime(metricData.CreateTime).
				SetUpdateTime(metricData.UpdateTime).
				SetProjectID(metricData.ProjectID).
				SetCollectedAt(metricData.CollectedAt)

			if metricData.MetricDescriptorJSON != nil {
				update.SetMetricDescriptorJSON(metricData.MetricDescriptorJSON)
			}
			if metricData.LabelExtractorsJSON != nil {
				update.SetLabelExtractorsJSON(metricData.LabelExtractorsJSON)
			}
			if metricData.BucketOptionsJSON != nil {
				update.SetBucketOptionsJSON(metricData.BucketOptionsJSON)
			}

			_, err := update.Save(ctx)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update log metric %s: %w", metricData.Name, err)
			}
		}

		if diff.IsNew {
			if err := s.history.CreateHistory(ctx, tx, metricData, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create history for log metric %s: %w", metricData.Name, err)
			}
		} else if diff.IsChanged {
			if err := s.history.UpdateHistory(ctx, tx, existing, metricData, diff, now); err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to update history for log metric %s: %w", metricData.Name, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// DeleteStaleLogMetrics removes log metrics that were not collected in the latest run.
func (s *Service) DeleteStaleLogMetrics(ctx context.Context, projectID string, collectedAt time.Time) error {
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

	stale, err := tx.BronzeGCPLoggingLogMetric.Query().
		Where(
			bronzegcplogginglogmetric.ProjectID(projectID),
			bronzegcplogginglogmetric.CollectedAtLT(collectedAt),
		).
		All(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	for _, staleMetric := range stale {
		if err := s.history.CloseHistory(ctx, tx, staleMetric.ID, now); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to close history for log metric %s: %w", staleMetric.ID, err)
		}
		if err := tx.BronzeGCPLoggingLogMetric.DeleteOne(staleMetric).Exec(ctx); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to delete log metric %s: %w", staleMetric.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
