package neg

import (
	"reflect"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

type NegDiff struct {
	IsNew     bool
	IsChanged bool
}

func (d *NegDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged
}

func DiffNegData(old *ent.BronzeGCPComputeNeg, new *NegData) *NegDiff {
	diff := &NegDiff{}

	if old == nil {
		diff.IsNew = true
		return diff
	}

	if old.Name != new.Name ||
		old.Description != new.Description ||
		old.CreationTimestamp != new.CreationTimestamp ||
		old.SelfLink != new.SelfLink ||
		old.Network != new.Network ||
		old.Subnetwork != new.Subnetwork ||
		old.Zone != new.Zone ||
		old.NetworkEndpointType != new.NetworkEndpointType ||
		old.DefaultPort != new.DefaultPort ||
		old.Size != new.Size ||
		old.Region != new.Region ||
		!reflect.DeepEqual(old.AnnotationsJSON, new.AnnotationsJSON) ||
		!reflect.DeepEqual(old.AppEngineJSON, new.AppEngineJSON) ||
		!reflect.DeepEqual(old.CloudFunctionJSON, new.CloudFunctionJSON) ||
		!reflect.DeepEqual(old.CloudRunJSON, new.CloudRunJSON) ||
		!reflect.DeepEqual(old.PscDataJSON, new.PscDataJSON) {
		diff.IsChanged = true
	}

	return diff
}
