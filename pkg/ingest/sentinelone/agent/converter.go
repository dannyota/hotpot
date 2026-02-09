package agent

import (
	"encoding/json"
	"fmt"
	"time"
)

// AgentData holds converted agent data ready for Ent insertion.
type AgentData struct {
	ResourceID              string
	ComputerName            string
	ExternalIP              string
	SiteName                string
	AccountID               string
	AccountName             string
	AgentVersion            string
	OSType                  string
	OSName                  string
	OSRevision              string
	OSArch                  string
	IsActive                bool
	IsInfected              bool
	IsDecommissioned        bool
	MachineType             string
	Domain                  string
	UUID                    string
	NetworkStatus           string
	LastActiveDate          *time.Time
	RegisteredAt            *time.Time
	APIUpdatedAt            *time.Time
	OSStartTime             *time.Time
	ThreatCount             int
	EncryptedApplications   bool
	GroupName               string
	GroupID                 string
	CPUCount                int
	CoreCount               int
	CPUId                   string
	TotalMemory             int64
	ModelName               string
	SerialNumber            string
	StorageEncryptionStatus string
	NetworkInterfacesJSON   json.RawMessage
	CollectedAt             time.Time

	// Child data
	NICs []NICData
}

// NICData holds converted NIC data.
type NICData struct {
	InterfaceID string
	Name        string
	Description string
	Type        string
	InetJSON    json.RawMessage
	Inet6JSON   json.RawMessage
	Physical    string
	GatewayIP   string
	GatewayMac  string
}

// ConvertAgent converts an API agent to AgentData.
func ConvertAgent(agent APIAgent, collectedAt time.Time) (*AgentData, error) {
	data := &AgentData{
		ResourceID:              agent.ID,
		ComputerName:            agent.ComputerName,
		ExternalIP:              agent.ExternalIP,
		SiteName:                agent.SiteName,
		AccountID:               agent.AccountID,
		AccountName:             agent.AccountName,
		AgentVersion:            agent.AgentVersion,
		OSType:                  agent.OSType,
		OSName:                  agent.OSName,
		OSRevision:              agent.OSRevision,
		OSArch:                  agent.OSArch,
		IsActive:                agent.IsActive,
		IsInfected:              agent.IsInfected,
		IsDecommissioned:        agent.IsDecommissioned,
		MachineType:             agent.MachineType,
		Domain:                  agent.Domain,
		UUID:                    agent.UUID,
		NetworkStatus:           agent.NetworkStatus,
		LastActiveDate:          agent.LastActiveDate,
		RegisteredAt:            agent.RegisteredAt,
		APIUpdatedAt:            agent.UpdatedAt,
		OSStartTime:             agent.OSStartTime,
		ThreatCount:             agent.ActiveThreats,
		EncryptedApplications:   agent.EncryptedApplications,
		GroupName:               agent.GroupName,
		GroupID:                 agent.GroupID,
		CPUCount:                agent.CPUCount,
		CoreCount:               agent.CoreCount,
		CPUId:                   agent.CPUId,
		TotalMemory:             agent.TotalMemory,
		ModelName:               agent.ModelName,
		SerialNumber:            agent.SerialNumber,
		StorageEncryptionStatus: agent.StorageEncryptionStatus,
		CollectedAt:             collectedAt,
	}

	// Store full network interfaces as JSONB snapshot
	if len(agent.NetworkInterfaces) > 0 {
		nicsJSON, err := json.Marshal(agent.NetworkInterfaces)
		if err != nil {
			return nil, fmt.Errorf("marshal network interfaces for agent %s: %w", agent.ID, err)
		}
		data.NetworkInterfacesJSON = nicsJSON
	}

	// Convert NICs to child data
	data.NICs = ConvertNICs(agent.NetworkInterfaces)

	return data, nil
}

// ConvertNICs converts API network interfaces to NICData.
func ConvertNICs(nics []APINetworkInterface) []NICData {
	if len(nics) == 0 {
		return nil
	}

	result := make([]NICData, 0, len(nics))
	for _, nic := range nics {
		n := NICData{
			InterfaceID: nic.ID,
			Name:        nic.Name,
			Description: nic.Description,
			Type:        nic.Type,
			Physical:    nic.Physical,
			GatewayIP:   nic.GatewayIP,
			GatewayMac:  nic.GatewayMac,
		}

		if len(nic.Inet) > 0 {
			n.InetJSON, _ = json.Marshal(nic.Inet)
		}
		if len(nic.Inet6) > 0 {
			n.Inet6JSON, _ = json.Marshal(nic.Inet6)
		}

		result = append(result, n)
	}

	return result
}
