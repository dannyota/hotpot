package app

import "github.com/dannyota/hotpot/pkg/storage/ent"

// AppDiff represents changes between old and new app states.
type AppDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffAppData compares old Ent entity and new data.
func DiffAppData(old *ent.BronzeS1App, new *AppData) *AppDiff {
	if old == nil {
		return &AppDiff{IsNew: true}
	}

	changed := old.Name != new.Name ||
		old.Publisher != new.Publisher ||
		old.Version != new.Version ||
		old.Size != new.Size ||
		old.AppType != new.AppType ||
		old.OsType != new.OsType ||
		old.AgentID != new.AgentID ||
		old.AgentComputerName != new.AgentComputerName ||
		old.AgentMachineType != new.AgentMachineType ||
		old.AgentIsActive != new.AgentIsActive ||
		old.AgentIsDecommissioned != new.AgentIsDecommissioned ||
		old.RiskLevel != new.RiskLevel ||
		old.Signed != new.Signed

	return &AppDiff{IsChanged: changed}
}
