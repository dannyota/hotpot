package revision

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/run/apiv2/runpb"
)

// RevisionData holds converted Cloud Run revision data ready for Ent insertion.
type RevisionData struct {
	ID                            string
	Name                          string
	UID                           string
	Generation                    int64
	LabelsJSON                    json.RawMessage
	AnnotationsJSON               json.RawMessage
	CreateTime                    string
	UpdateTime                    string
	DeleteTime                    string
	LaunchStage                   int
	ServiceName                   string
	ScalingJSON                   json.RawMessage
	ContainersJSON                json.RawMessage
	VolumesJSON                   json.RawMessage
	ExecutionEnvironment          int
	EncryptionKey                 string
	MaxInstanceRequestConcurrency int
	Timeout                       string
	ServiceAccount                string
	Reconciling                   bool
	ConditionsJSON                json.RawMessage
	ObservedGeneration            int64
	LogURI                        string
	Etag                          string
	ProjectID                     string
	Location                      string
	CollectedAt                   time.Time
}

// ConvertRevision converts a raw GCP API Cloud Run revision to Ent-compatible data.
func ConvertRevision(serviceName string, rev *runpb.Revision, projectID string, collectedAt time.Time) (*RevisionData, error) {
	data := &RevisionData{
		ID:                            rev.GetName(),
		Name:                          extractRevisionShortName(rev.GetName()),
		UID:                           rev.GetUid(),
		Generation:                    rev.GetGeneration(),
		LaunchStage:                   int(rev.GetLaunchStage()),
		ServiceName:                   serviceName,
		ExecutionEnvironment:          int(rev.GetExecutionEnvironment()),
		EncryptionKey:                 rev.GetEncryptionKey(),
		MaxInstanceRequestConcurrency: int(rev.GetMaxInstanceRequestConcurrency()),
		ServiceAccount:                rev.GetServiceAccount(),
		Reconciling:                   rev.GetReconciling(),
		ObservedGeneration:            rev.GetObservedGeneration(),
		LogURI:                        rev.GetLogUri(),
		Etag:                          rev.GetEtag(),
		ProjectID:                     projectID,
		Location:                      extractLocation(rev.GetName()),
		CollectedAt:                   collectedAt,
	}

	if rev.GetTimeout() != nil {
		data.Timeout = fmt.Sprintf("%ds", rev.GetTimeout().GetSeconds())
	}

	if rev.GetCreateTime() != nil {
		data.CreateTime = rev.GetCreateTime().AsTime().Format(time.RFC3339)
	}
	if rev.GetUpdateTime() != nil {
		data.UpdateTime = rev.GetUpdateTime().AsTime().Format(time.RFC3339)
	}
	if rev.GetDeleteTime() != nil {
		data.DeleteTime = rev.GetDeleteTime().AsTime().Format(time.RFC3339)
	}

	if labels := rev.GetLabels(); len(labels) > 0 {
		b, err := json.Marshal(labels)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels for revision %s: %w", rev.GetName(), err)
		}
		data.LabelsJSON = b
	}

	if annotations := rev.GetAnnotations(); len(annotations) > 0 {
		b, err := json.Marshal(annotations)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal annotations for revision %s: %w", rev.GetName(), err)
		}
		data.AnnotationsJSON = b
	}

	if scaling := rev.GetScaling(); scaling != nil {
		b, err := json.Marshal(scaling)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal scaling for revision %s: %w", rev.GetName(), err)
		}
		data.ScalingJSON = b
	}

	if containers := rev.GetContainers(); len(containers) > 0 {
		b, err := json.Marshal(containers)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal containers for revision %s: %w", rev.GetName(), err)
		}
		data.ContainersJSON = b
	}

	if volumes := rev.GetVolumes(); len(volumes) > 0 {
		b, err := json.Marshal(volumes)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal volumes for revision %s: %w", rev.GetName(), err)
		}
		data.VolumesJSON = b
	}

	if conditions := rev.GetConditions(); len(conditions) > 0 {
		b, err := json.Marshal(conditions)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal conditions for revision %s: %w", rev.GetName(), err)
		}
		data.ConditionsJSON = b
	}

	return data, nil
}

// extractRevisionShortName extracts the revision short name from a resource name.
// Format: projects/{project}/locations/{location}/services/{service}/revisions/{revision}
func extractRevisionShortName(resourceName string) string {
	parts := strings.Split(resourceName, "/")
	if len(parts) >= 8 {
		return parts[7]
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
