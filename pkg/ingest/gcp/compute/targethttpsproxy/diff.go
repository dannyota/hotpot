package targethttpsproxy

import (
	"encoding/json"

	"hotpot/pkg/storage/ent"
)

// TargetHttpsProxyDiff represents changes between old and new target HTTPS proxy states.
type TargetHttpsProxyDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffTargetHttpsProxyData compares existing Ent entity with new TargetHttpsProxyData and returns differences.
func DiffTargetHttpsProxyData(old *ent.BronzeGCPComputeTargetHttpsProxy, new *TargetHttpsProxyData) *TargetHttpsProxyDiff {
	diff := &TargetHttpsProxyDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.SelfLink != new.SelfLink ||
		old.Fingerprint != new.Fingerprint ||
		old.URLMap != new.UrlMap ||
		old.QuicOverride != new.QuicOverride ||
		old.ServerTLSPolicy != new.ServerTlsPolicy ||
		old.AuthorizationPolicy != new.AuthorizationPolicy ||
		old.CertificateMap != new.CertificateMap ||
		old.SslPolicy != new.SslPolicy ||
		old.TLSEarlyData != new.TlsEarlyData ||
		old.ProxyBind != new.ProxyBind ||
		old.HTTPKeepAliveTimeoutSec != new.HttpKeepAliveTimeoutSec ||
		old.Region != new.Region ||
		jsonChanged(old.SslCertificatesJSON, new.SslCertificatesJSON) {
		diff.IsChanged = true
	}

	return diff
}

func jsonChanged(old, new []interface{}) bool {
	oldBytes, _ := json.Marshal(old)
	newBytes, _ := json.Marshal(new)
	return string(oldBytes) != string(newBytes)
}

// HasAnyChange returns true if any part of the target HTTPS proxy changed.
func (d *TargetHttpsProxyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
