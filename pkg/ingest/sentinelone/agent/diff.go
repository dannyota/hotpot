package agent

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// AgentDiff represents changes between old and new agent states.
type AgentDiff struct {
	IsNew     bool
	IsChanged bool
	NICsDiff  ChildDiff
}

// ChildDiff represents changes in a child collection.
type ChildDiff struct {
	Changed bool
}

// HasAnyChange returns true if any part of the agent changed.
func (d *AgentDiff) HasAnyChange() bool {
	return d.IsNew || d.IsChanged || d.NICsDiff.Changed
}

// DiffAgentData compares old Ent entity and new data.
func DiffAgentData(old *ent.BronzeS1Agent, new *AgentData) *AgentDiff {
	if old == nil {
		return &AgentDiff{
			IsNew:    true,
			NICsDiff: ChildDiff{Changed: true},
		}
	}

	diff := &AgentDiff{}

	diff.IsChanged = hasAgentFieldsChanged(old, new)
	diff.NICsDiff = diffNICsData(old.Edges.Nics, new.NICs)

	return diff
}

func hasAgentFieldsChanged(old *ent.BronzeS1Agent, new *AgentData) bool {
	return old.ComputerName != new.ComputerName ||
		old.ExternalIP != new.ExternalIP ||
		old.SiteName != new.SiteName ||
		old.AccountID != new.AccountID ||
		old.AccountName != new.AccountName ||
		old.AgentVersion != new.AgentVersion ||
		old.OsType != new.OSType ||
		old.OsName != new.OSName ||
		old.OsRevision != new.OSRevision ||
		old.OsArch != new.OSArch ||
		old.IsActive != new.IsActive ||
		old.IsInfected != new.IsInfected ||
		old.IsDecommissioned != new.IsDecommissioned ||
		old.MachineType != new.MachineType ||
		old.Domain != new.Domain ||
		old.UUID != new.UUID ||
		old.NetworkStatus != new.NetworkStatus ||
		old.ThreatCount != new.ThreatCount ||
		old.EncryptedApplications != new.EncryptedApplications ||
		old.GroupName != new.GroupName ||
		old.GroupID != new.GroupID ||
		old.CPUCount != new.CPUCount ||
		old.CoreCount != new.CoreCount ||
		old.CPUID != new.CPUId ||
		old.TotalMemory != new.TotalMemory ||
		old.ModelName != new.ModelName ||
		old.SerialNumber != new.SerialNumber ||
		old.StorageEncryptionStatus != new.StorageEncryptionStatus ||
		!bytes.Equal(old.NetworkInterfacesJSON, new.NetworkInterfacesJSON)
}

func diffNICsData(old []*ent.BronzeS1AgentNIC, new []NICData) ChildDiff {
	if len(old) != len(new) {
		return ChildDiff{Changed: true}
	}
	for i := range old {
		if hasNICChangedData(old[i], &new[i]) {
			return ChildDiff{Changed: true}
		}
	}
	return ChildDiff{Changed: false}
}

func hasNICChangedData(old *ent.BronzeS1AgentNIC, new *NICData) bool {
	return old.InterfaceID != new.InterfaceID ||
		old.Name != new.Name ||
		old.Description != new.Description ||
		old.Type != new.Type ||
		old.Physical != new.Physical ||
		old.GatewayIP != new.GatewayIP ||
		old.GatewayMAC != new.GatewayMac ||
		!bytes.Equal(old.InetJSON, new.InetJSON) ||
		!bytes.Equal(old.Inet6JSON, new.Inet6JSON)
}
