package service

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/run/apiv2/runpb"
)

// ServiceData holds converted Cloud Run service data ready for Ent insertion.
type ServiceData struct {
	ID                    string
	Name                  string
	Description           string
	UID                   string
	Generation            int64
	LabelsJSON            json.RawMessage
	AnnotationsJSON       json.RawMessage
	CreateTime            string
	UpdateTime            string
	DeleteTime            string
	Creator               string
	LastModifier          string
	Ingress               int
	LaunchStage           int
	TemplateJSON          json.RawMessage
	TrafficJSON           json.RawMessage
	URI                   string
	ObservedGeneration    int64
	TerminalConditionJSON json.RawMessage
	ConditionsJSON        json.RawMessage
	LatestReadyRevision   string
	LatestCreatedRevision string
	TrafficStatusesJSON   json.RawMessage
	Reconciling           bool
	Etag                  string
	ProjectID             string
	Location              string
	CollectedAt           time.Time
}

// ConvertService converts a raw GCP API Cloud Run service to Ent-compatible data.
func ConvertService(svc *runpb.Service, projectID string, collectedAt time.Time) (*ServiceData, error) {
	data := &ServiceData{
		ID:                 svc.GetName(),
		Name:               extractServiceShortName(svc.GetName()),
		UID:                svc.GetUid(),
		Generation:         svc.GetGeneration(),
		Description:        svc.GetDescription(),
		Creator:            svc.GetCreator(),
		LastModifier:       svc.GetLastModifier(),
		Ingress:            int(svc.GetIngress()),
		LaunchStage:        int(svc.GetLaunchStage()),
		URI:                svc.GetUri(),
		ObservedGeneration: svc.GetObservedGeneration(),
		LatestReadyRevision:   svc.GetLatestReadyRevision(),
		LatestCreatedRevision: svc.GetLatestCreatedRevision(),
		Reconciling:           svc.GetReconciling(),
		Etag:                  svc.GetEtag(),
		ProjectID:             projectID,
		Location:              extractLocation(svc.GetName()),
		CollectedAt:           collectedAt,
	}

	if svc.GetCreateTime() != nil {
		data.CreateTime = svc.GetCreateTime().AsTime().Format(time.RFC3339)
	}
	if svc.GetUpdateTime() != nil {
		data.UpdateTime = svc.GetUpdateTime().AsTime().Format(time.RFC3339)
	}
	if svc.GetDeleteTime() != nil {
		data.DeleteTime = svc.GetDeleteTime().AsTime().Format(time.RFC3339)
	}

	if labels := svc.GetLabels(); len(labels) > 0 {
		b, err := json.Marshal(labels)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels for service %s: %w", svc.GetName(), err)
		}
		data.LabelsJSON = b
	}

	if annotations := svc.GetAnnotations(); len(annotations) > 0 {
		b, err := json.Marshal(annotations)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal annotations for service %s: %w", svc.GetName(), err)
		}
		data.AnnotationsJSON = b
	}

	if tmpl := svc.GetTemplate(); tmpl != nil {
		b, err := json.Marshal(tmpl)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal template for service %s: %w", svc.GetName(), err)
		}
		data.TemplateJSON = b
	}

	if traffic := svc.GetTraffic(); len(traffic) > 0 {
		b, err := json.Marshal(traffic)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal traffic for service %s: %w", svc.GetName(), err)
		}
		data.TrafficJSON = b
	}

	if tc := svc.GetTerminalCondition(); tc != nil {
		b, err := json.Marshal(tc)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal terminal_condition for service %s: %w", svc.GetName(), err)
		}
		data.TerminalConditionJSON = b
	}

	if conditions := svc.GetConditions(); len(conditions) > 0 {
		b, err := json.Marshal(conditions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal conditions for service %s: %w", svc.GetName(), err)
		}
		data.ConditionsJSON = b
	}

	if statuses := svc.GetTrafficStatuses(); len(statuses) > 0 {
		b, err := json.Marshal(statuses)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal traffic_statuses for service %s: %w", svc.GetName(), err)
		}
		data.TrafficStatusesJSON = b
	}

	return data, nil
}

// extractServiceShortName extracts the service short name from a resource name.
// Format: projects/{project}/locations/{location}/services/{service}
func extractServiceShortName(resourceName string) string {
	parts := strings.Split(resourceName, "/")
	if len(parts) >= 6 {
		return parts[5]
	}
	return resourceName
}

// extractLocation extracts the location from a resource name.
// Format: projects/{project}/locations/{location}/...
func extractLocation(resourceName string) string {
	parts := strings.Split(resourceName, "/")
	if len(parts) >= 4 {
		return parts[3]
	}
	return ""
}
