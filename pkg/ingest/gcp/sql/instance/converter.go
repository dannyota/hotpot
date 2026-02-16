package instance

import (
	"encoding/json"
	"fmt"
	"time"

	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

// InstanceData holds converted Cloud SQL instance data ready for Ent insertion.
type InstanceData struct {
	ResourceID                       string
	Name                             string
	DatabaseVersion                  string
	State                            string
	Region                           string
	GceZone                          string
	SecondaryGceZone                 string
	InstanceType                     string
	ConnectionName                   string
	ServiceAccountEmailAddress       string
	SelfLink                         string
	SettingsJSON                     json.RawMessage
	ServerCaCertJSON                 json.RawMessage
	IpAddressesJSON                  json.RawMessage
	ReplicaConfigurationJSON         json.RawMessage
	FailoverReplicaJSON              json.RawMessage
	DiskEncryptionConfigurationJSON  json.RawMessage
	DiskEncryptionStatusJSON         json.RawMessage
	ProjectID                        string
	CollectedAt                      time.Time

	// Child data
	Labels []LabelData
}

// LabelData holds a key-value label pair.
type LabelData struct {
	Key   string
	Value string
}

// ConvertInstance converts a GCP API DatabaseInstance to InstanceData.
// Preserves raw API data with minimal transformation.
func ConvertInstance(inst *sqladmin.DatabaseInstance, projectID string, collectedAt time.Time) (*InstanceData, error) {
	data := &InstanceData{
		ResourceID:                 inst.Name,
		Name:                       inst.Name,
		DatabaseVersion:            inst.DatabaseVersion,
		State:                      inst.State,
		Region:                     inst.Region,
		GceZone:                    inst.GceZone,
		SecondaryGceZone:           inst.SecondaryGceZone,
		InstanceType:               inst.InstanceType,
		ConnectionName:             inst.ConnectionName,
		ServiceAccountEmailAddress: inst.ServiceAccountEmailAddress,
		SelfLink:                   inst.SelfLink,
		ProjectID:                  projectID,
		CollectedAt:                collectedAt,
	}

	// Convert Settings to JSONB (nil -> SQL NULL, data -> JSON bytes)
	if inst.Settings != nil {
		var err error
		data.SettingsJSON, err = json.Marshal(inst.Settings)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal settings for instance %s: %w", inst.Name, err)
		}
	}

	// Convert ServerCaCert to JSONB
	if inst.ServerCaCert != nil {
		var err error
		data.ServerCaCertJSON, err = json.Marshal(inst.ServerCaCert)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal server_ca_cert for instance %s: %w", inst.Name, err)
		}
	}

	// Convert IpAddresses to JSONB
	if inst.IpAddresses != nil {
		var err error
		data.IpAddressesJSON, err = json.Marshal(inst.IpAddresses)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ip_addresses for instance %s: %w", inst.Name, err)
		}
	}

	// Convert ReplicaConfiguration to JSONB
	if inst.ReplicaConfiguration != nil {
		var err error
		data.ReplicaConfigurationJSON, err = json.Marshal(inst.ReplicaConfiguration)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal replica_configuration for instance %s: %w", inst.Name, err)
		}
	}

	// Convert FailoverReplica to JSONB
	if inst.FailoverReplica != nil {
		var err error
		data.FailoverReplicaJSON, err = json.Marshal(inst.FailoverReplica)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal failover_replica for instance %s: %w", inst.Name, err)
		}
	}

	// Convert DiskEncryptionConfiguration to JSONB
	if inst.DiskEncryptionConfiguration != nil {
		var err error
		data.DiskEncryptionConfigurationJSON, err = json.Marshal(inst.DiskEncryptionConfiguration)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal disk_encryption_configuration for instance %s: %w", inst.Name, err)
		}
	}

	// Convert DiskEncryptionStatus to JSONB
	if inst.DiskEncryptionStatus != nil {
		var err error
		data.DiskEncryptionStatusJSON, err = json.Marshal(inst.DiskEncryptionStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal disk_encryption_status for instance %s: %w", inst.Name, err)
		}
	}

	// Convert UserLabels map to LabelData slice
	data.Labels = ConvertLabels(inst.Settings)

	return data, nil
}

// ConvertLabels converts instance user labels from Settings to LabelData.
func ConvertLabels(settings *sqladmin.Settings) []LabelData {
	if settings == nil || len(settings.UserLabels) == 0 {
		return nil
	}

	result := make([]LabelData, 0, len(settings.UserLabels))
	for key, value := range settings.UserLabels {
		result = append(result, LabelData{
			Key:   key,
			Value: value,
		})
	}

	return result
}
