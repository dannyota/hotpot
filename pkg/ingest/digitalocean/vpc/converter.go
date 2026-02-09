package vpc

import (
	"time"

	"github.com/digitalocean/godo"
)

// VpcData holds converted VPC data ready for Ent insertion.
type VpcData struct {
	ResourceID   string
	Name         string
	Description  string
	Region       string
	IPRange      string
	URN          string
	IsDefault    bool
	APICreatedAt *time.Time
	CollectedAt  time.Time
}

// ConvertVpc converts a godo VPC to VpcData.
func ConvertVpc(v *godo.VPC, collectedAt time.Time) *VpcData {
	data := &VpcData{
		ResourceID:  v.ID,
		Name:        v.Name,
		Description: v.Description,
		Region:      v.RegionSlug,
		IPRange:     v.IPRange,
		URN:         v.URN,
		IsDefault:   v.Default,
		CollectedAt: collectedAt,
	}

	if !v.CreatedAt.IsZero() {
		t := v.CreatedAt
		data.APICreatedAt = &t
	}

	return data
}
