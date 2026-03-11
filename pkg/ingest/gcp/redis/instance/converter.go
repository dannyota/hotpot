package instance

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/redis/apiv1/redispb"
)

// InstanceData holds converted Redis instance data ready for Ent insertion.
type InstanceData struct {
	ID                               string
	Name                             string
	DisplayName                      string
	LabelsJSON                       json.RawMessage
	LocationID                       string
	AlternativeLocationID            string
	RedisVersion                     string
	ReservedIPRange                  string
	SecondaryIPRange                 string
	Host                             string
	Port                             int32
	CurrentLocationID                string
	CreateTime                       string
	State                            int32
	StatusMessage                    string
	RedisConfigsJSON                 json.RawMessage
	Tier                             int32
	MemorySizeGB                     int32
	AuthorizedNetwork                string
	PersistenceIamIdentity           string
	ConnectMode                      int32
	AuthEnabled                      bool
	ServerCaCertsJSON                json.RawMessage
	TransitEncryptionMode            int32
	MaintenancePolicyJSON            json.RawMessage
	MaintenanceScheduleJSON          json.RawMessage
	ReplicaCount                     int32
	NodesJSON                        json.RawMessage
	ReadEndpoint                     string
	ReadEndpointPort                 int32
	ReadReplicasMode                 int32
	CustomerManagedKey               string
	PersistenceConfigJSON            json.RawMessage
	SuspensionReasonsJSON            json.RawMessage
	MaintenanceVersion               string
	AvailableMaintenanceVersionsJSON json.RawMessage
	ProjectID                        string
	CollectedAt                      time.Time
}

// ConvertInstance converts a GCP API Redis Instance to Ent-compatible data.
func ConvertInstance(inst *redispb.Instance, projectID string, collectedAt time.Time) (*InstanceData, error) {
	data := &InstanceData{
		ID:                    inst.GetName(),
		Name:                  inst.GetName(),
		DisplayName:           inst.GetDisplayName(),
		LocationID:            inst.GetLocationId(),
		AlternativeLocationID: inst.GetAlternativeLocationId(),
		RedisVersion:          inst.GetRedisVersion(),
		ReservedIPRange:       inst.GetReservedIpRange(),
		SecondaryIPRange:      inst.GetSecondaryIpRange(),
		Host:                  inst.GetHost(),
		Port:                  inst.GetPort(),
		CurrentLocationID:     inst.GetCurrentLocationId(),
		State:                 int32(inst.GetState()),
		StatusMessage:         inst.GetStatusMessage(),
		Tier:                  int32(inst.GetTier()),
		MemorySizeGB:          inst.GetMemorySizeGb(),
		AuthorizedNetwork:     inst.GetAuthorizedNetwork(),
		PersistenceIamIdentity: inst.GetPersistenceIamIdentity(),
		ConnectMode:           int32(inst.GetConnectMode()),
		AuthEnabled:           inst.GetAuthEnabled(),
		TransitEncryptionMode: int32(inst.GetTransitEncryptionMode()),
		ReplicaCount:          inst.GetReplicaCount(),
		ReadEndpoint:          inst.GetReadEndpoint(),
		ReadEndpointPort:      inst.GetReadEndpointPort(),
		ReadReplicasMode:      int32(inst.GetReadReplicasMode()),
		CustomerManagedKey:    inst.GetCustomerManagedKey(),
		MaintenanceVersion:    inst.GetMaintenanceVersion(),
		ProjectID:             projectID,
		CollectedAt:           collectedAt,
	}

	if inst.GetCreateTime() != nil {
		data.CreateTime = inst.GetCreateTime().AsTime().Format(time.RFC3339)
	}

	if len(inst.GetLabels()) > 0 {
		j, err := json.Marshal(inst.GetLabels())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal labels for instance %s: %w", inst.GetName(), err)
		}
		data.LabelsJSON = j
	}

	if len(inst.GetRedisConfigs()) > 0 {
		j, err := json.Marshal(inst.GetRedisConfigs())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal redis_configs for instance %s: %w", inst.GetName(), err)
		}
		data.RedisConfigsJSON = j
	}

	if len(inst.GetServerCaCerts()) > 0 {
		j, err := json.Marshal(inst.GetServerCaCerts())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal server_ca_certs for instance %s: %w", inst.GetName(), err)
		}
		data.ServerCaCertsJSON = j
	}

	if inst.GetMaintenancePolicy() != nil {
		j, err := json.Marshal(inst.GetMaintenancePolicy())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal maintenance_policy for instance %s: %w", inst.GetName(), err)
		}
		data.MaintenancePolicyJSON = j
	}

	if inst.GetMaintenanceSchedule() != nil {
		j, err := json.Marshal(inst.GetMaintenanceSchedule())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal maintenance_schedule for instance %s: %w", inst.GetName(), err)
		}
		data.MaintenanceScheduleJSON = j
	}

	if len(inst.GetNodes()) > 0 {
		j, err := json.Marshal(inst.GetNodes())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal nodes for instance %s: %w", inst.GetName(), err)
		}
		data.NodesJSON = j
	}

	if inst.GetPersistenceConfig() != nil {
		j, err := json.Marshal(inst.GetPersistenceConfig())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal persistence_config for instance %s: %w", inst.GetName(), err)
		}
		data.PersistenceConfigJSON = j
	}

	if len(inst.GetSuspensionReasons()) > 0 {
		j, err := json.Marshal(inst.GetSuspensionReasons())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal suspension_reasons for instance %s: %w", inst.GetName(), err)
		}
		data.SuspensionReasonsJSON = j
	}

	if len(inst.GetAvailableMaintenanceVersions()) > 0 {
		j, err := json.Marshal(inst.GetAvailableMaintenanceVersions())
		if err != nil {
			return nil, fmt.Errorf("failed to marshal available_maintenance_versions for instance %s: %w", inst.GetName(), err)
		}
		data.AvailableMaintenanceVersionsJSON = j
	}

	return data, nil
}
