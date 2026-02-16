package loadbalancer

import (
	"encoding/json"
	"time"

	"github.com/digitalocean/godo"
)

// LoadBalancerData holds converted Load Balancer data ready for Ent insertion.
type LoadBalancerData struct {
	ResourceID                   string
	Name                         string
	IP                           string
	Ipv6                         string
	SizeSlug                     string
	SizeUnit                     uint32
	LbType                       string
	Algorithm                    string
	Status                       string
	Region                       string
	Tag                          string
	RedirectHTTPToHTTPS          bool
	EnableProxyProtocol          bool
	EnableBackendKeepalive       bool
	VpcUUID                      string
	ProjectID                    string
	HTTPIdleTimeoutSeconds       *uint64
	DisableLetsEncryptDNSRecords *bool
	Network                      string
	NetworkStack                 string
	TLSCipherPolicy              string
	APICreatedAt                 string
	ForwardingRulesJSON          json.RawMessage
	HealthCheckJSON              json.RawMessage
	StickySessionsJSON           json.RawMessage
	FirewallJSON                 json.RawMessage
	DomainsJSON                  json.RawMessage
	GlbSettingsJSON              json.RawMessage
	DropletIdsJSON               json.RawMessage
	TagsJSON                     json.RawMessage
	TargetLoadBalancerIdsJSON    json.RawMessage
	CollectedAt                  time.Time
}

// ConvertLoadBalancer converts a godo LoadBalancer to LoadBalancerData.
func ConvertLoadBalancer(v godo.LoadBalancer, collectedAt time.Time) *LoadBalancerData {
	data := &LoadBalancerData{
		ResourceID:                   v.ID,
		Name:                         v.Name,
		IP:                           v.IP,
		Ipv6:                         v.IPv6,
		SizeSlug:                     v.SizeSlug,
		SizeUnit:                     v.SizeUnit,
		LbType:                       v.Type,
		Algorithm:                    v.Algorithm,
		Status:                       v.Status,
		Tag:                          v.Tag,
		RedirectHTTPToHTTPS:          v.RedirectHttpToHttps,
		EnableProxyProtocol:          v.EnableProxyProtocol,
		EnableBackendKeepalive:       v.EnableBackendKeepalive,
		VpcUUID:                      v.VPCUUID,
		ProjectID:                    v.ProjectID,
		HTTPIdleTimeoutSeconds:       v.HTTPIdleTimeoutSeconds,
		DisableLetsEncryptDNSRecords: v.DisableLetsEncryptDNSRecords,
		Network:                      v.Network,
		NetworkStack:                 v.NetworkStack,
		TLSCipherPolicy:              v.TLSCipherPolicy,
		APICreatedAt:                 v.Created,
		CollectedAt:                  collectedAt,
	}

	if v.Region != nil {
		data.Region = v.Region.Slug
	}

	if len(v.ForwardingRules) > 0 {
		data.ForwardingRulesJSON, _ = json.Marshal(v.ForwardingRules)
	}

	if v.HealthCheck != nil {
		data.HealthCheckJSON, _ = json.Marshal(v.HealthCheck)
	}

	if v.StickySessions != nil {
		data.StickySessionsJSON, _ = json.Marshal(v.StickySessions)
	}

	if v.Firewall != nil {
		data.FirewallJSON, _ = json.Marshal(v.Firewall)
	}

	if len(v.Domains) > 0 {
		data.DomainsJSON, _ = json.Marshal(v.Domains)
	}

	if v.GLBSettings != nil {
		data.GlbSettingsJSON, _ = json.Marshal(v.GLBSettings)
	}

	if len(v.DropletIDs) > 0 {
		data.DropletIdsJSON, _ = json.Marshal(v.DropletIDs)
	}

	if len(v.Tags) > 0 {
		data.TagsJSON, _ = json.Marshal(v.Tags)
	}

	if len(v.TargetLoadBalancerIDs) > 0 {
		data.TargetLoadBalancerIdsJSON, _ = json.Marshal(v.TargetLoadBalancerIDs)
	}

	return data
}
