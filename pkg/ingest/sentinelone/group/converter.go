package group

import "time"

// GroupData holds converted group data ready for Ent insertion.
type GroupData struct {
	ResourceID   string
	Name         string
	SiteID       string
	Type         string
	IsDefault    bool
	Inherits     bool
	Rank         *int
	TotalAgents  int
	Creator      string
	CreatorID    string
	FilterName   string
	FilterID     string
	APICreatedAt *time.Time
	APIUpdatedAt *time.Time
	CollectedAt  time.Time
}

// ConvertGroup converts an API group to GroupData.
func ConvertGroup(g APIGroup, collectedAt time.Time) *GroupData {
	data := &GroupData{
		ResourceID:  g.ID,
		Name:        g.Name,
		SiteID:      g.SiteID,
		Type:        g.Type,
		IsDefault:   g.IsDefault,
		Inherits:    g.Inherits,
		Rank:        g.Rank,
		TotalAgents: g.TotalAgents,
		Creator:     g.Creator,
		CreatorID:   g.CreatorID,
		FilterName:  g.FilterName,
		FilterID:    g.FilterID,
		CollectedAt: collectedAt,
	}

	if g.CreatedAt != nil {
		if t, err := time.Parse(time.RFC3339, *g.CreatedAt); err == nil {
			data.APICreatedAt = &t
		}
	}
	if g.UpdatedAt != nil {
		if t, err := time.Parse(time.RFC3339, *g.UpdatedAt); err == nil {
			data.APIUpdatedAt = &t
		}
	}

	return data
}
