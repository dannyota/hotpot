package threat

import (
	"bytes"

	"hotpot/pkg/storage/ent"
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
			!bytes.Equal(old.ThreatInfoJSON, new.ThreatInfoJSON),
	}
}
