package firewall

import (
	"encoding/json"
	"time"

	"github.com/digitalocean/godo"
)

// FirewallData holds converted Firewall data ready for Ent insertion.
type FirewallData struct {
	ResourceID         string
	Name               string
	Status             string
	InboundRulesJSON   json.RawMessage
	OutboundRulesJSON  json.RawMessage
	DropletIdsJSON     json.RawMessage
	TagsJSON           json.RawMessage
	APICreatedAt       string
	PendingChangesJSON json.RawMessage
	CollectedAt        time.Time
}

// ConvertFirewall converts a godo Firewall to FirewallData.
func ConvertFirewall(v godo.Firewall, collectedAt time.Time) *FirewallData {
	data := &FirewallData{
		ResourceID:   v.ID,
		Name:         v.Name,
		Status:       v.Status,
		APICreatedAt: v.Created,
		CollectedAt:  collectedAt,
	}

	if len(v.InboundRules) > 0 {
		data.InboundRulesJSON, _ = json.Marshal(v.InboundRules)
	}

	if len(v.OutboundRules) > 0 {
		data.OutboundRulesJSON, _ = json.Marshal(v.OutboundRules)
	}

	if len(v.DropletIDs) > 0 {
		data.DropletIdsJSON, _ = json.Marshal(v.DropletIDs)
	}

	if len(v.Tags) > 0 {
		data.TagsJSON, _ = json.Marshal(v.Tags)
	}

	if len(v.PendingChanges) > 0 {
		data.PendingChangesJSON, _ = json.Marshal(v.PendingChanges)
	}

	return data
}
