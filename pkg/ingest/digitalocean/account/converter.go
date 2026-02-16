package account

import (
	"time"

	"github.com/digitalocean/godo"
)

// AccountData holds converted Account data ready for Ent insertion.
type AccountData struct {
	ResourceID     string
	Email          string
	Name           string
	Status         string
	StatusMessage  string
	DropletLimit   int
	FloatingIPLimit int
	ReservedIPLimit int
	VolumeLimit    int
	EmailVerified  bool
	TeamName       string
	TeamUUID       string
	CollectedAt    time.Time
}

// ConvertAccount converts a godo Account to AccountData.
func ConvertAccount(v *godo.Account, collectedAt time.Time) *AccountData {
	data := &AccountData{
		ResourceID:      v.UUID,
		Email:           v.Email,
		Name:            v.Name,
		Status:          v.Status,
		StatusMessage:   v.StatusMessage,
		DropletLimit:    v.DropletLimit,
		FloatingIPLimit: v.FloatingIPLimit,
		ReservedIPLimit: v.ReservedIPLimit,
		VolumeLimit:     v.VolumeLimit,
		EmailVerified:   v.EmailVerified,
		CollectedAt:     collectedAt,
	}

	if v.Team != nil {
		data.TeamName = v.Team.Name
		data.TeamUUID = v.Team.UUID
	}

	return data
}
