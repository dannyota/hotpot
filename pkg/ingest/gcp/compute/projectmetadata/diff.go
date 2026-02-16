package projectmetadata

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ProjectMetadataDiff represents changes between old and new project metadata states.
type ProjectMetadataDiff struct {
	IsNew     bool
	IsChanged bool

	// Child diff (for granular tracking)
	ItemsDiff ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// DiffProjectMetadataData compares old Ent entity and new data.
func DiffProjectMetadataData(old *ent.BronzeGCPComputeProjectMetadata, new *ProjectMetadataData) *ProjectMetadataDiff {
	if old == nil {
		return &ProjectMetadataDiff{
			IsNew:     true,
			ItemsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &ProjectMetadataDiff{}

	// Compare metadata-level fields
	diff.IsChanged = hasMetadataFieldsChanged(old, new)

	// Compare items children
	var oldItems []*ent.BronzeGCPComputeProjectMetadataItem
	if old.Edges.Items != nil {
		oldItems = old.Edges.Items
	}
	diff.ItemsDiff = diffItems(oldItems, new.Items)

	return diff
}

// HasAnyChange returns true if any part of the metadata changed.
func (d *ProjectMetadataDiff) HasAnyChange() bool {
	if d.IsNew || d.IsChanged {
		return true
	}
	return d.ItemsDiff.Changed
}

// hasMetadataFieldsChanged compares metadata-level fields (excluding children).
func hasMetadataFieldsChanged(old *ent.BronzeGCPComputeProjectMetadata, new *ProjectMetadataData) bool {
	return old.Name != new.Name ||
		old.DefaultServiceAccount != new.DefaultServiceAccount ||
		old.DefaultNetworkTier != new.DefaultNetworkTier ||
		old.XpnProjectStatus != new.XpnProjectStatus ||
		old.CreationTimestamp != new.CreationTimestamp ||
		!bytes.Equal(old.UsageExportLocationJSON, new.UsageExportLocationJSON)
}

func diffItems(old []*ent.BronzeGCPComputeProjectMetadataItem, new []ItemData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}

	// Build map of old items by key+value
	type itemKey struct {
		key   string
		value string
	}
	oldMap := make(map[itemKey]bool)
	for _, item := range old {
		k := itemKey{key: item.Key, value: item.Value}
		oldMap[k] = true
	}

	// Compare each new item
	for _, newItem := range new {
		k := itemKey{key: newItem.Key, value: newItem.Value}
		if !oldMap[k] {
			return ChildDiff{Changed: true}
		}
	}

	return ChildDiff{Changed: false}
}
