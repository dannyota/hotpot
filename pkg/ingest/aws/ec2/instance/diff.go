package instance

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// InstanceDiff represents changes between old and new instance states.
type InstanceDiff struct {
	IsNew     bool
	IsChanged bool

	// Child diffs
	TagsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffInstanceData compares old Ent entity and new data.
func DiffInstanceData(old *ent.BronzeAWSEC2Instance, new *InstanceData) *InstanceDiff {
	if old == nil {
		return &InstanceDiff{
			IsNew:    true,
			TagsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &InstanceDiff{}

	// Compare instance-level fields
	diff.IsChanged = hasInstanceFieldsChanged(old, new)

	// Compare tags
	diff.TagsDiff = diffTagsData(old.Edges.Tags, new.Tags)

	return diff
}

// HasAnyChange returns true if any part of the instance changed.
func (d *InstanceDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.TagsDiff.Changed
}

func hasInstanceFieldsChanged(old *ent.BronzeAWSEC2Instance, new *InstanceData) bool {
	return old.Name != new.Name ||
		old.InstanceType != new.InstanceType ||
		old.State != new.State ||
		old.VpcID != new.VpcID ||
		old.SubnetID != new.SubnetID ||
		old.PrivateIPAddress != new.PrivateIPAddress ||
		old.PublicIPAddress != new.PublicIPAddress ||
		old.AmiID != new.AmiID ||
		old.KeyName != new.KeyName ||
		old.Platform != new.Platform ||
		old.Architecture != new.Architecture ||
		!bytes.Equal(old.SecurityGroupsJSON, new.SecurityGroupJSON)
}

func diffTagsData(old []*ent.BronzeAWSEC2InstanceTag, new []TagData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	oldMap := make(map[string]string)
	for _, t := range old {
		oldMap[t.Key] = t.Value
	}
	for _, t := range new {
		if v, ok := oldMap[t.Key]; !ok || v != t.Value {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}
