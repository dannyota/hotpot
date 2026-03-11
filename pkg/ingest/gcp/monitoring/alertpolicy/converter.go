package alertpolicy

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/protobuf/encoding/protojson"
)

// AlertPolicyData holds converted alert policy data ready for Ent insertion.
type AlertPolicyData struct {
	ID                       string
	Name                     string
	DisplayName              string
	DocumentationJSON        json.RawMessage
	UserLabelsJSON           json.RawMessage
	ConditionsJSON           json.RawMessage
	Combiner                 int
	Enabled                  bool
	NotificationChannelsJSON json.RawMessage
	CreationRecordJSON       json.RawMessage
	MutationRecordJSON       json.RawMessage
	AlertStrategyJSON        json.RawMessage
	Severity                 int
	ProjectID                string
	CollectedAt              time.Time
}

// ConvertAlertPolicy converts a raw GCP API alert policy to Ent-compatible data.
func ConvertAlertPolicy(policy *monitoringpb.AlertPolicy, projectID string, collectedAt time.Time) *AlertPolicyData {
	marshaler := protojson.MarshalOptions{UseProtoNames: true}

	data := &AlertPolicyData{
		ID:          policy.GetName(),
		Name:        policy.GetName(),
		DisplayName: policy.GetDisplayName(),
		Combiner:    int(policy.GetCombiner()),
		Severity:    int(policy.GetSeverity()),
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	if policy.GetEnabled() != nil {
		data.Enabled = policy.GetEnabled().GetValue()
	}

	if doc := policy.GetDocumentation(); doc != nil {
		if b, err := marshaler.Marshal(doc); err == nil {
			data.DocumentationJSON = b
		}
	}

	if labels := policy.GetUserLabels(); len(labels) > 0 {
		if b, err := json.Marshal(labels); err == nil {
			data.UserLabelsJSON = b
		}
	}

	if conditions := policy.GetConditions(); len(conditions) > 0 {
		if b, err := marshalConditions(marshaler, conditions); err == nil {
			data.ConditionsJSON = b
		}
	}

	if channels := policy.GetNotificationChannels(); len(channels) > 0 {
		if b, err := json.Marshal(channels); err == nil {
			data.NotificationChannelsJSON = b
		}
	}

	if cr := policy.GetCreationRecord(); cr != nil {
		if b, err := marshaler.Marshal(cr); err == nil {
			data.CreationRecordJSON = b
		}
	}

	if mr := policy.GetMutationRecord(); mr != nil {
		if b, err := marshaler.Marshal(mr); err == nil {
			data.MutationRecordJSON = b
		}
	}

	if as := policy.GetAlertStrategy(); as != nil {
		if b, err := marshaler.Marshal(as); err == nil {
			data.AlertStrategyJSON = b
		}
	}

	return data
}

// marshalConditions marshals a slice of Condition protos to JSON.
func marshalConditions(marshaler protojson.MarshalOptions, conditions []*monitoringpb.AlertPolicy_Condition) (json.RawMessage, error) {
	var items []json.RawMessage
	for _, c := range conditions {
		b, err := marshaler.Marshal(c)
		if err != nil {
			return nil, err
		}
		items = append(items, b)
	}
	return json.Marshal(items)
}
