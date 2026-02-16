package secret

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// SecretDiff represents changes between old and new secret states.
type SecretDiff struct {
	IsNew     bool
	IsChanged bool

	LabelDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffSecretData compares old Ent entity and new data.
func DiffSecretData(old *ent.BronzeGCPSecretManagerSecret, new *SecretData) *SecretDiff {
	if old == nil {
		return &SecretDiff{
			IsNew:     true,
			LabelDiff: ChildDiff{Changed: true},
		}
	}

	diff := &SecretDiff{}

	diff.IsChanged = hasSecretFieldsChanged(old, new)

	var oldLabels []*ent.BronzeGCPSecretManagerSecretLabel
	if old.Edges.Labels != nil {
		oldLabels = old.Edges.Labels
	}
	diff.LabelDiff = diffLabels(oldLabels, new.Labels)

	return diff
}

// HasAnyChange returns true if any part of the secret changed.
func (d *SecretDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.LabelDiff.Changed
}

func hasSecretFieldsChanged(old *ent.BronzeGCPSecretManagerSecret, new *SecretData) bool {
	return old.Name != new.Name ||
		old.CreateTime != new.CreateTime ||
		old.Etag != new.Etag ||
		!bytes.Equal(old.ReplicationJSON, new.ReplicationJSON) ||
		!bytes.Equal(old.RotationJSON, new.RotationJSON) ||
		!bytes.Equal(old.TopicsJSON, new.TopicsJSON) ||
		!bytes.Equal(old.VersionAliasesJSON, new.VersionAliasesJSON) ||
		!bytes.Equal(old.AnnotationsJSON, new.AnnotationsJSON)
}

func diffLabels(old []*ent.BronzeGCPSecretManagerSecretLabel, new []LabelData) ChildDiff {
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
