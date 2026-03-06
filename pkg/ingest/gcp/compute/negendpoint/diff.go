package negendpoint

import (
	"reflect"

	entcompute "danny.vn/hotpot/pkg/storage/ent/gcp/compute"
)

type NegEndpointDiff struct {
	IsNew     bool
	IsChanged bool
}

func (d *NegEndpointDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

func DiffNegEndpointData(old *entcompute.BronzeGCPComputeNegEndpoint, new *NegEndpointData) *NegEndpointDiff {
	diff := &NegEndpointDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Instance != new.Instance ||
		old.IPAddress != new.IpAddress ||
		old.Ipv6Address != new.Ipv6Address ||
		old.Port != new.Port ||
		old.Fqdn != new.Fqdn ||
		old.NegName != new.NegName ||
		old.Zone != new.Zone ||
		!reflect.DeepEqual(old.AnnotationsJSON, new.AnnotationsJSON) {
		diff.IsChanged = true
	}

	return diff
}
