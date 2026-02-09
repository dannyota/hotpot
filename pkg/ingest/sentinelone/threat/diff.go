package threat

import (
	"bytes"

	"github.com/dannyota/hotpot/pkg/storage/ent"
)

// ThreatDiff represents changes between old and new threat states.
type ThreatDiff struct {
	IsNew     bool
	IsChanged bool
}

// DiffThreatData compares old Ent entity and new data.
func DiffThreatData(old *ent.BronzeS1Threat, new *ThreatData) *ThreatDiff {
	if old == nil {
		return &ThreatDiff{IsNew: true}
	}

	return &ThreatDiff{
		IsChanged: old.AgentID != new.AgentID ||
			old.Classification != new.Classification ||
			old.ThreatName != new.ThreatName ||
			old.FilePath != new.FilePath ||
			old.Status != new.Status ||
			old.AnalystVerdict != new.AnalystVerdict ||
			old.ConfidenceLevel != new.ConfidenceLevel ||
			old.InitiatedBy != new.InitiatedBy ||
			!bytes.Equal(old.ThreatInfoJSON, new.ThreatInfoJSON) ||
			old.FileContentHash != new.FileContentHash ||
			old.FileSha256 != new.FileSHA256 ||
			old.CloudVerdict != new.CloudVerdict ||
			old.ClassificationSource != new.ClassificationSource ||
			old.SiteID != new.SiteID ||
			old.SiteName != new.SiteName ||
			old.AccountID != new.AccountID ||
			old.AccountName != new.AccountName ||
			old.AgentComputerName != new.AgentComputerName ||
			old.AgentOsType != new.AgentOsType ||
			old.AgentMachineType != new.AgentMachineType ||
			old.AgentIsActive != new.AgentIsActive ||
			old.AgentIsDecommissioned != new.AgentIsDecommissioned ||
			old.AgentVersion != new.AgentVersion,
	}
}
