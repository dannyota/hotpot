package bucket

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// BucketDiff represents changes between old and new bucket states.
type BucketDiff struct {
	IsNew     bool
	IsChanged bool

	LabelDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffBucketData compares old Ent entity and new data.
func DiffBucketData(old *ent.BronzeGCPStorageBucket, new *BucketData) *BucketDiff {
	if old == nil {
		return &BucketDiff{
			IsNew:     true,
			LabelDiff: ChildDiff{Changed: true},
		}
	}

	diff := &BucketDiff{}

	diff.IsChanged = hasBucketFieldsChanged(old, new)

	var oldLabels []*ent.BronzeGCPStorageBucketLabel
	if old.Edges.Labels != nil {
		oldLabels = old.Edges.Labels
	}
	diff.LabelDiff = diffLabels(oldLabels, new.Labels)

	return diff
}

// HasAnyChange returns true if any part of the bucket changed.
func (d *BucketDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.LabelDiff.Changed
}

func hasBucketFieldsChanged(old *ent.BronzeGCPStorageBucket, new *BucketData) bool {
	return old.Name != new.Name ||
		old.Location != new.Location ||
		old.StorageClass != new.StorageClass ||
		old.ProjectNumber != new.ProjectNumber ||
		old.TimeCreated != new.TimeCreated ||
		old.Updated != new.Updated ||
		old.DefaultEventBasedHold != new.DefaultEventBasedHold ||
		old.Metageneration != new.Metageneration ||
		old.Etag != new.Etag ||
		!bytes.Equal(old.IamConfigurationJSON, new.IamConfigurationJSON) ||
		!bytes.Equal(old.EncryptionJSON, new.EncryptionJSON) ||
		!bytes.Equal(old.LifecycleJSON, new.LifecycleJSON) ||
		!bytes.Equal(old.VersioningJSON, new.VersioningJSON) ||
		!bytes.Equal(old.RetentionPolicyJSON, new.RetentionPolicyJSON) ||
		!bytes.Equal(old.LoggingJSON, new.LoggingJSON) ||
		!bytes.Equal(old.CorsJSON, new.CorsJSON) ||
		!bytes.Equal(old.WebsiteJSON, new.WebsiteJSON) ||
		!bytes.Equal(old.AutoclassJSON, new.AutoclassJSON)
}

func diffLabels(old []*ent.BronzeGCPStorageBucketLabel, new []LabelData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}

	oldMap := make(map[string]string)
	for _, l := range old {
		oldMap[l.Key] = l.Value
	}

	for _, newL := range new {
		if v, ok := oldMap[newL.Key]; !ok || v != newL.Value {
			return ChildDiff{Changed: true}
		}
	}

	return ChildDiff{Changed: false}
}
