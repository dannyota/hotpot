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
	APIUpdatedAt          *time.Time
	FileContentHash       string
	FileSHA256            string
	CloudVerdict          string
	ClassificationSource  string
	SiteID                string
	SiteName              string
	AccountID             string
	AccountName           string
	AgentComputerName     string
	AgentOsType           string
	AgentMachineType      string
	AgentIsActive         bool
	AgentIsDecommissioned bool
	AgentVersion          string
	CollectedAt     time.Time
}

// ConvertThreat converts an API threat to ThreatData.
func ConvertThreat(t APIThreat, collectedAt time.Time) *ThreatData {
	agentID := t.AgentRealtimeInfo.AgentID

	data := &ThreatData{
		ResourceID:            t.ID,
		AgentID:               agentID,
		Classification:        t.Classification,
		ThreatName:            t.ThreatName,
		FilePath:              t.FilePath,
		Status:                t.MitigationStatus,
		AnalystVerdict:        t.AnalystVerdict,
		ConfidenceLevel:       t.ConfidenceLevel,
		InitiatedBy:           t.InitiatedBy,
		APICreatedAt:          t.CreatedAt,
		ThreatInfoJSON:        t.ThreatInfo,
		APIUpdatedAt:          t.UpdatedAt,
		FileContentHash:       t.FileContentHash,
		CloudVerdict:          t.CloudVerdict,
		ClassificationSource:  t.ClassificationSource,
		SiteID:                t.AgentRealtimeInfo.SiteID,
		SiteName:              t.AgentRealtimeInfo.SiteName,
		AccountID:             t.AgentRealtimeInfo.AccountID,
		AccountName:           t.AgentRealtimeInfo.AccountName,
		AgentComputerName:     t.AgentRealtimeInfo.AgentComputerName,
		AgentOsType:           t.AgentRealtimeInfo.AgentOsType,
		AgentMachineType:      t.AgentRealtimeInfo.AgentMachineType,
		AgentIsActive:         t.AgentRealtimeInfo.AgentIsActive,
		AgentIsDecommissioned: t.AgentRealtimeInfo.AgentIsDecommissioned,
		AgentVersion:          t.AgentRealtimeInfo.AgentVersion,
		CollectedAt:           collectedAt,
	}

	// Extract file_sha256 from threatInfo JSON
	if t.ThreatInfo != nil {
		var info ThreatInfoData
		if err := json.Unmarshal(t.ThreatInfo, &info); err == nil {
			data.FileSHA256 = info.SHA256
		}
	}

	return data
}
