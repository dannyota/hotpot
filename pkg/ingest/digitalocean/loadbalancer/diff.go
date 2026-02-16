package loadbalancer

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// LoadBalancerDiff represents changes between old and new Load Balancer states.
type LoadBalancerDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffLoadBalancerData compares old Ent entity and new data.
func DiffLoadBalancerData(old *ent.BronzeDOLoadBalancer, new *LoadBalancerData) *LoadBalancerDiff {
	if old == nil {
		return &LoadBalancerDiff{IsNew: true}
	}

	changed := old.Name != new.Name ||
		old.IP != new.IP ||
		old.Ipv6 != new.Ipv6 ||
		old.SizeSlug != new.SizeSlug ||
		old.SizeUnit != new.SizeUnit ||
		old.LbType != new.LbType ||
		old.Algorithm != new.Algorithm ||
		old.Status != new.Status ||
		old.Region != new.Region ||
		old.Tag != new.Tag ||
		old.RedirectHTTPToHTTPS != new.RedirectHTTPToHTTPS ||
		old.EnableProxyProtocol != new.EnableProxyProtocol ||
		old.EnableBackendKeepalive != new.EnableBackendKeepalive ||
		old.VpcUUID != new.VpcUUID ||
		old.ProjectID != new.ProjectID ||
		!ptrUint64Equal(old.HTTPIdleTimeoutSeconds, new.HTTPIdleTimeoutSeconds) ||
		!ptrBoolEqual(old.DisableLetsEncryptDNSRecords, new.DisableLetsEncryptDNSRecords) ||
		old.Network != new.Network ||
		old.NetworkStack != new.NetworkStack ||
		old.TLSCipherPolicy != new.TLSCipherPolicy ||
		old.APICreatedAt != new.APICreatedAt ||
		!bytes.Equal(old.ForwardingRulesJSON, new.ForwardingRulesJSON) ||
		!bytes.Equal(old.HealthCheckJSON, new.HealthCheckJSON) ||
		!bytes.Equal(old.StickySessionsJSON, new.StickySessionsJSON) ||
		!bytes.Equal(old.FirewallJSON, new.FirewallJSON) ||
		!bytes.Equal(old.DomainsJSON, new.DomainsJSON) ||
		!bytes.Equal(old.GlbSettingsJSON, new.GlbSettingsJSON) ||
		!bytes.Equal(old.DropletIdsJSON, new.DropletIdsJSON) ||
		!bytes.Equal(old.TagsJSON, new.TagsJSON) ||
		!bytes.Equal(old.TargetLoadBalancerIdsJSON, new.TargetLoadBalancerIdsJSON)

	return &LoadBalancerDiff{IsChanged: changed}
}

func ptrUint64Equal(a, b *uint64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrBoolEqual(a, b *bool) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}
