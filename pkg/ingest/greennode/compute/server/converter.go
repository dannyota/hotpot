package server

import (
	"encoding/json"
	"fmt"
	"time"

	computev2 "danny.vn/greennode/services/compute/v2"
)

// ServerData represents a converted server ready for Ent insertion.
type ServerData struct {
	ID               string
	Name             string
	Status           string
	Location         string
	ZoneID           string
	CreatedAtAPI     string
	BootVolumeID     string
	EncryptionVolume bool
	Licence          bool
	Metadata         string
	MigrateState     string
	Product          string
	ServerGroupID    string
	ServerGroupName  string
	SSHKeyName       string
	StopBeforeMigrate bool
	User             string
	ImageID          string
	ImageType        string
	ImageVersion     string
	FlavorID         string
	FlavorName       string
	FlavorCPU        int64
	FlavorMemory     int64
	FlavorGPU        int64
	FlavorBandwidth  int64
	InterfacesJSON   json.RawMessage
	Region           string
	ProjectID        string
	CollectedAt      time.Time

	SecGroups []SecGroupData
}

// SecGroupData represents a security group attached to a server.
type SecGroupData struct {
	UUID string
	Name string
}

// ConvertServer converts a GreenNode SDK Server to ServerData.
func ConvertServer(s *computev2.Server, projectID, region string, collectedAt time.Time) (*ServerData, error) {
	data := &ServerData{
		ID:                s.Uuid,
		Name:              s.Name,
		Status:            s.Status,
		Location:          s.Location,
		ZoneID:            s.ZoneID,
		CreatedAtAPI:      s.CreatedAt,
		BootVolumeID:      s.BootVolumeID,
		EncryptionVolume:  s.EncryptionVolume,
		Licence:           s.Licence,
		Metadata:          s.Metadata,
		MigrateState:      s.MigrateState,
		Product:           s.Product,
		ServerGroupID:     s.ServerGroupID,
		ServerGroupName:   s.ServerGroupName,
		SSHKeyName:        s.SshKeyName,
		StopBeforeMigrate: s.StopBeforeMigrate,
		User:              s.User,
		ImageID:           s.Image.ID,
		ImageType:         s.Image.ImageType,
		ImageVersion:      s.Image.ImageVersion,
		FlavorID:          s.Flavor.FlavorID,
		FlavorName:        s.Flavor.Name,
		FlavorCPU:         s.Flavor.Cpu,
		FlavorMemory:      s.Flavor.Memory,
		FlavorGPU:         s.Flavor.Gpu,
		FlavorBandwidth:   s.Flavor.Bandwidth,
		Region:            region,
		ProjectID:         projectID,
		CollectedAt:       collectedAt,
	}

	// Marshal network interfaces to JSON
	if len(s.ExternalInterfaces) > 0 || len(s.InternalInterfaces) > 0 {
		ifaces := map[string]interface{}{
			"external": s.ExternalInterfaces,
			"internal": s.InternalInterfaces,
		}
		ifacesJSON, err := json.Marshal(ifaces)
		if err != nil {
			return nil, fmt.Errorf("marshal interfaces for server %s: %w", s.Name, err)
		}
		data.InterfacesJSON = ifacesJSON
	}

	// Convert security groups
	data.SecGroups = ConvertSecGroups(s.SecGroups)

	return data, nil
}

// ConvertSecGroups converts SDK security groups to SecGroupData.
func ConvertSecGroups(secGroups []computev2.ServerSecgroup) []SecGroupData {
	if len(secGroups) == 0 {
		return nil
	}
	result := make([]SecGroupData, 0, len(secGroups))
	for _, sg := range secGroups {
		result = append(result, SecGroupData{
			UUID: sg.Uuid,
			Name: sg.Name,
		})
	}
	return result
}
