package uptimecheck

import (
	"encoding/json"
	"time"

	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	"google.golang.org/protobuf/encoding/protojson"
)

// UptimeCheckData holds converted uptime check config data ready for Ent insertion.
type UptimeCheckData struct {
	ID                   string
	Name                 string
	DisplayName          string
	MonitoredResourceJSON json.RawMessage
	ResourceGroupJSON    json.RawMessage
	HttpCheckJSON        json.RawMessage
	TcpCheckJSON         json.RawMessage
	Period               string
	Timeout              string
	ContentMatchersJSON  json.RawMessage
	CheckerType          int
	SelectedRegionsJSON  json.RawMessage
	IsInternal           bool
	InternalCheckersJSON json.RawMessage
	UserLabelsJSON       json.RawMessage
	ProjectID            string
	CollectedAt          time.Time
}

// ConvertUptimeCheckConfig converts a raw GCP API uptime check config to Ent-compatible data.
func ConvertUptimeCheckConfig(cfg *monitoringpb.UptimeCheckConfig, projectID string, collectedAt time.Time) *UptimeCheckData {
	marshaler := protojson.MarshalOptions{UseProtoNames: true}

	data := &UptimeCheckData{
		ID:          cfg.GetName(),
		Name:        cfg.GetName(),
		DisplayName: cfg.GetDisplayName(),
		CheckerType: int(cfg.GetCheckerType()),
		IsInternal:  cfg.GetIsInternal(),
		ProjectID:   projectID,
		CollectedAt: collectedAt,
	}

	if period := cfg.GetPeriod(); period != nil {
		data.Period = period.String()
	}

	if timeout := cfg.GetTimeout(); timeout != nil {
		data.Timeout = timeout.String()
	}

	if mr := cfg.GetMonitoredResource(); mr != nil {
		if b, err := marshaler.Marshal(mr); err == nil {
			data.MonitoredResourceJSON = b
		}
	}

	if rg := cfg.GetResourceGroup(); rg != nil {
		if b, err := marshaler.Marshal(rg); err == nil {
			data.ResourceGroupJSON = b
		}
	}

	if hc := cfg.GetHttpCheck(); hc != nil {
		if b, err := marshaler.Marshal(hc); err == nil {
			data.HttpCheckJSON = b
		}
	}

	if tc := cfg.GetTcpCheck(); tc != nil {
		if b, err := marshaler.Marshal(tc); err == nil {
			data.TcpCheckJSON = b
		}
	}

	if cms := cfg.GetContentMatchers(); len(cms) > 0 {
		if b, err := marshalContentMatchers(marshaler, cms); err == nil {
			data.ContentMatchersJSON = b
		}
	}

	if regions := cfg.GetSelectedRegions(); len(regions) > 0 {
		regionInts := make([]int, len(regions))
		for i, r := range regions {
			regionInts[i] = int(r)
		}
		if b, err := json.Marshal(regionInts); err == nil {
			data.SelectedRegionsJSON = b
		}
	}

	if ics := cfg.GetInternalCheckers(); len(ics) > 0 {
		if b, err := marshalInternalCheckers(marshaler, ics); err == nil {
			data.InternalCheckersJSON = b
		}
	}

	if labels := cfg.GetUserLabels(); len(labels) > 0 {
		if b, err := json.Marshal(labels); err == nil {
			data.UserLabelsJSON = b
		}
	}

	return data
}

// marshalContentMatchers marshals a slice of ContentMatcher protos to JSON.
func marshalContentMatchers(marshaler protojson.MarshalOptions, matchers []*monitoringpb.UptimeCheckConfig_ContentMatcher) (json.RawMessage, error) {
	var items []json.RawMessage
	for _, m := range matchers {
		b, err := marshaler.Marshal(m)
		if err != nil {
			return nil, err
		}
		items = append(items, b)
	}
	return json.Marshal(items)
}

// marshalInternalCheckers marshals a slice of InternalChecker protos to JSON.
func marshalInternalCheckers(marshaler protojson.MarshalOptions, checkers []*monitoringpb.InternalChecker) (json.RawMessage, error) {
	var items []json.RawMessage
	for _, c := range checkers {
		b, err := marshaler.Marshal(c)
		if err != nil {
			return nil, err
		}
		items = append(items, b)
	}
	return json.Marshal(items)
}
