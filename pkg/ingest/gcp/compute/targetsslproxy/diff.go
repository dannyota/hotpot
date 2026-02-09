package targetsslproxy

import (
	"encoding/json"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// TargetSslProxyDiff represents changes between old and new target SSL proxy states.
type TargetSslProxyDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffTargetSslProxyData compares existing Ent entity with new TargetSslProxyData.
func DiffTargetSslProxyData(old *ent.BronzeGCPComputeTargetSslProxy, new *TargetSslProxyData) *TargetSslProxyDiff {
	diff := &TargetSslProxyDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.SelfLink != new.SelfLink ||
		old.Service != new.Service ||
		old.ProxyHeader != new.ProxyHeader ||
		old.CertificateMap != new.CertificateMap ||
		old.SslPolicy != new.SslPolicy ||
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

// HasAnyChange returns true if any part of the target SSL proxy changed.
func (d *TargetSslProxyDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}
