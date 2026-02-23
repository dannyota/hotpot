package sshkey

import (
	"time"

	computev2 "danny.vn/greennode/services/compute/v2"
)

// SSHKeyData represents a converted SSH key ready for Ent insertion.
type SSHKeyData struct {
	ID           string
	Name         string
	CreatedAtAPI string
	PubKey       string
	Status       string
	Region       string
	ProjectID    string
	CollectedAt  time.Time
}

// ConvertSSHKey converts a GreenNode SDK SSHKey to SSHKeyData.
func ConvertSSHKey(k *computev2.SSHKey, projectID, region string, collectedAt time.Time) *SSHKeyData {
	return &SSHKeyData{
		ID:           k.ID,
		Name:         k.Name,
		CreatedAtAPI: k.CreatedAt,
		PubKey:       k.PubKey,
		Status:       k.Status,
		Region:       region,
		ProjectID:    projectID,
		CollectedAt:  collectedAt,
	}
}
