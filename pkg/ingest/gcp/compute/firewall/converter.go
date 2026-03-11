package firewall

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// FirewallData holds converted firewall data ready for Ent insertion.
type FirewallData struct {
	ID                        string
	Name                      string
	Description               string
	SelfLink                  string
	CreationTimestamp         string
	Network                   string
	Priority                  int32
	Direction                 string
	Disabled                  bool
	SourceRangesJSON          json.RawMessage
	DestinationRangesJSON     json.RawMessage
	SourceTagsJSON            json.RawMessage
	TargetTagsJSON            json.RawMessage
	SourceServiceAccountsJSON json.RawMessage
	TargetServiceAccountsJSON json.RawMessage
	LogConfigJSON             json.RawMessage
	Allowed                   []AllowedData
	Denied                    []DeniedData
	ProjectID                 string
	CollectedAt               time.Time
}

// AllowedData holds converted allowed rule data.
type AllowedData struct {
	IpProtocol string
	PortsJSON  json.RawMessage
}

// DeniedData holds converted denied rule data.
type DeniedData struct {
	IpProtocol string
	PortsJSON  json.RawMessage
}

// ConvertFirewall converts a GCP API Firewall to Ent-compatible data.
// Preserves raw API data with minimal transformation.
func ConvertFirewall(fw *computepb.Firewall, projectID string, collectedAt time.Time) (*FirewallData, error) {
	data := &FirewallData{
		ID:                fmt.Sprintf("%d", fw.GetId()),
		Name:              fw.GetName(),
		Description:       fw.GetDescription(),
		SelfLink:          fw.GetSelfLink(),
		CreationTimestamp: fw.GetCreationTimestamp(),
		Network:           fw.GetNetwork(),
		Priority:          int32(fw.GetPriority()),
		Direction:         fw.GetDirection(),
		Disabled:          fw.GetDisabled(),
		ProjectID:         projectID,
		CollectedAt:       collectedAt,
	}

	// The proto GetId() returns uint64, convert to string
	data.ID = fmt.Sprintf("%d", fw.GetId())

	// Convert slice fields to JSONB (nil -> SQL NULL, data -> JSON bytes)
	if fw.GetSourceRanges() != nil {
		var err error
		data.SourceRangesJSON, err = json.Marshal(fw.GetSourceRanges())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal source_ranges for firewall %s: %w", fw.GetName(), err)
		}
	}

	if fw.GetDestinationRanges() != nil {
		var err error
		data.DestinationRangesJSON, err = json.Marshal(fw.GetDestinationRanges())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal destination_ranges for firewall %s: %w", fw.GetName(), err)
		}
	}

	if fw.GetSourceTags() != nil {
		var err error
		data.SourceTagsJSON, err = json.Marshal(fw.GetSourceTags())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal source_tags for firewall %s: %w", fw.GetName(), err)
		}
	}

	if fw.GetTargetTags() != nil {
		var err error
		data.TargetTagsJSON, err = json.Marshal(fw.GetTargetTags())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal target_tags for firewall %s: %w", fw.GetName(), err)
		}
	}

	if fw.GetSourceServiceAccounts() != nil {
		var err error
		data.SourceServiceAccountsJSON, err = json.Marshal(fw.GetSourceServiceAccounts())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal source_service_accounts for firewall %s: %w", fw.GetName(), err)
		}
	}

	if fw.GetTargetServiceAccounts() != nil {
		var err error
		data.TargetServiceAccountsJSON, err = json.Marshal(fw.GetTargetServiceAccounts())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal target_service_accounts for firewall %s: %w", fw.GetName(), err)
		}
	}

	if fw.GetLogConfig() != nil {
		var err error
		data.LogConfigJSON, err = json.Marshal(fw.GetLogConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal log_config for firewall %s: %w", fw.GetName(), err)
		}
	}

	// Convert allowed/denied rules to child data
	data.Allowed = ConvertAllowed(fw.GetAllowed())
	data.Denied = ConvertDenied(fw.GetDenied())

	return data, nil
}

// ConvertAllowed converts firewall allowed rules from GCP API to data structs.
func ConvertAllowed(rules []*computepb.Allowed) []AllowedData {
	if len(rules) == 0 {
		return nil
	}

	result := make([]AllowedData, 0, len(rules))
	for _, rule := range rules {
		ad := AllowedData{
			IpProtocol: rule.GetIPProtocol(),
		}
		if rule.GetPorts() != nil {
			portsJSON, _ := json.Marshal(rule.GetPorts())
			ad.PortsJSON = portsJSON
		}
		result = append(result, ad)
	}

	return result
}

// ConvertDenied converts firewall denied rules from GCP API to data structs.
func ConvertDenied(rules []*computepb.Denied) []DeniedData {
	if len(rules) == 0 {
		return nil
	}

	result := make([]DeniedData, 0, len(rules))
	for _, rule := range rules {
		dd := DeniedData{
			IpProtocol: rule.GetIPProtocol(),
		}
		if rule.GetPorts() != nil {
			portsJSON, _ := json.Marshal(rule.GetPorts())
			dd.PortsJSON = portsJSON
		}
		result = append(result, dd)
	}

	return result
}
