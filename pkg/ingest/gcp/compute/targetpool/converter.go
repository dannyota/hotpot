package targetpool

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// TargetPoolData holds converted target pool data ready for Ent insertion.
type TargetPoolData struct {
	ID                string
	Name              string
	Description       string
	CreationTimestamp string
	SelfLink          string
	SessionAffinity   string
	BackupPool        string
	FailoverRatio     float32
	SecurityPolicy    string
	Region            string
	HealthChecksJSON  []interface{}
	InstancesJSON     []interface{}
	ProjectID         string
	CollectedAt       time.Time
}

// ConvertTargetPool converts a GCP API TargetPool to Ent-compatible data.
func ConvertTargetPool(tp *computepb.TargetPool, projectID string, collectedAt time.Time) (*TargetPoolData, error) {
	data := &TargetPoolData{
		ID:                fmt.Sprintf("%d", tp.GetId()),
		Name:              tp.GetName(),
		Description:       tp.GetDescription(),
		CreationTimestamp: tp.GetCreationTimestamp(),
		SelfLink:          tp.GetSelfLink(),
		SessionAffinity:   tp.GetSessionAffinity(),
		BackupPool:        tp.GetBackupPool(),
		FailoverRatio:     tp.GetFailoverRatio(),
		SecurityPolicy:    tp.GetSecurityPolicy(),
		Region:            tp.GetRegion(),
		ProjectID:         projectID,
		CollectedAt:       collectedAt,
	}

	if tp.HealthChecks != nil {
		b, err := json.Marshal(tp.HealthChecks)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal health checks for target pool %s: %w", tp.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.HealthChecksJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal health checks: %w", err)
		}
	}

	if tp.Instances != nil {
		b, err := json.Marshal(tp.Instances)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal instances for target pool %s: %w", tp.GetName(), err)
		}
		if err := json.Unmarshal(b, &data.InstancesJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal instances: %w", err)
		}
	}

	return data, nil
}
