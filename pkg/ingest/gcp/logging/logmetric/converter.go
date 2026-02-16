package logmetric

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/logging/apiv2/loggingpb"
)

// LogMetricData holds converted log metric data ready for Ent insertion.
type LogMetricData struct {
	ResourceID          string
	Name                string
	Description         string
	Filter              string
	MetricDescriptorJSON json.RawMessage
	LabelExtractorsJSON json.RawMessage
	BucketOptionsJSON   json.RawMessage
	ValueExtractor      string
	Version             string
	Disabled            bool
	CreateTime          string
	UpdateTime          string
	ProjectID           string
	CollectedAt         time.Time
}

// ConvertLogMetric converts a GCP API LogMetric to Ent-compatible data.
func ConvertLogMetric(m *loggingpb.LogMetric, projectID string, collectedAt time.Time) (*LogMetricData, error) {
	data := &LogMetricData{
		ResourceID:     m.GetName(),
		Name:           m.GetName(),
		Description:    m.GetDescription(),
		Filter:         m.GetFilter(),
		ValueExtractor: m.GetValueExtractor(),
		Version:        m.GetVersion().String(),
		Disabled:       m.GetDisabled(),
		ProjectID:      projectID,
		CollectedAt:    collectedAt,
	}

	if m.GetCreateTime() != nil {
		data.CreateTime = m.GetCreateTime().AsTime().Format(time.RFC3339)
	}

	if m.GetUpdateTime() != nil {
		data.UpdateTime = m.GetUpdateTime().AsTime().Format(time.RFC3339)
	}

	if m.GetMetricDescriptor() != nil {
		mdJSON, err := json.Marshal(m.GetMetricDescriptor())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal metric_descriptor for metric %s: %w", m.GetName(), err)
		}
		data.MetricDescriptorJSON = mdJSON
	}

	if m.GetLabelExtractors() != nil {
		leJSON, err := json.Marshal(m.GetLabelExtractors())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal label_extractors for metric %s: %w", m.GetName(), err)
		}
		data.LabelExtractorsJSON = leJSON
	}

	if m.GetBucketOptions() != nil {
		boJSON, err := json.Marshal(m.GetBucketOptions())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal bucket_options for metric %s: %w", m.GetName(), err)
		}
		data.BucketOptionsJSON = boJSON
	}

	return data, nil
}
