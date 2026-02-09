package threat

import (
	"encoding/json"
	"time"
)

// ThreatData holds converted threat data ready for Ent insertion.
type ThreatData struct {
	ResourceID      string
	AgentID         string
	Classification  string
	ThreatName      string
	FilePath        string
	Status          string
	AnalystVerdict  string
	ConfidenceLevel string
	InitiatedBy     string
	APICreatedAt    *time.Time
	ThreatInfoJSON  json.RawMessage
	CollectedAt     time.Time
}

// ConvertThreat converts an API threat to ThreatData.
func ConvertThreat(t APIThreat, collectedAt time.Time) *ThreatData {
	agentID := t.AgentRealtimeInfo.AgentID

	return &ThreatData{
		ResourceID:      t.ID,
		AgentID:         agentID,
		Classification:  t.Classification,
		ThreatName:      t.ThreatName,
		FilePath:        t.FilePath,
		Status:          t.MitigationStatus,
		AnalystVerdict:  t.AnalystVerdict,
		ConfidenceLevel: t.ConfidenceLevel,
		InitiatedBy:     t.InitiatedBy,
		APICreatedAt:    t.CreatedAt,
		ThreatInfoJSON:  t.ThreatInfo,
		CollectedAt:     collectedAt,
	}
}
