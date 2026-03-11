package network_discovery

import (
	"encoding/json"
	"time"
)

// NetworkDiscoveryData holds converted network discovery data ready for Ent insertion.
type NetworkDiscoveryData struct {
	ResourceID string

	// String fields
	Name                     string
	IPAddress                string
	Domain                   string
	SerialNumber             string
	Category                 string
	SubCategory              string
	ResourceType             string
	OS                       string
	OSFamily                 string
	OSVersion                string
	OSNameVersion            string
	Architecture             string
	Manufacturer             string
	CPU                      string
	MemoryReadable           string
	NetworkName              string
	AssetStatus              string
	AssetCriticality         string
	AssetEnvironment         string
	InfectionStatus          string
	DeviceReview             string
	EppUnsupportedUnknown    string
	AssetContactEmail        string
	LegacyIdentityPolicyName string
	PreviousOSType           string
	PreviousOSVersion        string
	PreviousDeviceFunction   string
	DetectedFromSite         string
	S1AccountID              string
	S1AccountName            string
	S1SiteID                 string
	S1SiteName               string
	S1GroupID                string
	S1GroupName              string
	S1ScopeID                string
	S1ScopeLevel             string
	S1ScopePath              string
	S1OnboardedAccountName   string
	S1OnboardedGroupName     string
	S1OnboardedSiteName      string
	S1OnboardedScopeLevel    string
	S1OnboardedScopePath     string

	// Int fields
	Memory               int
	CoreCount            int
	S1ManagementID       int
	S1ScopeType          int
	S1OnboardedAccountID int
	S1OnboardedGroupID   int
	S1OnboardedScopeID   int
	S1OnboardedSiteID    int

	// Bool fields
	IsAdConnector bool
	IsDcServer    bool
	AdsEnabled    bool

	// Time fields
	FirstSeenDt  *time.Time
	LastUpdateDt *time.Time
	LastActiveDt *time.Time
	LastRebootDt *time.Time
	S1UpdatedAt  *time.Time

	// JSON fields
	AgentJSON             json.RawMessage
	NetworkInterfacesJSON json.RawMessage
	AlertsJSON            json.RawMessage
	AlertsCountJSON       json.RawMessage
	DeviceReviewLogJSON   json.RawMessage
	IdentityJSON          json.RawMessage
	NotesJSON             json.RawMessage
	TagsJSON              json.RawMessage
	MissingCoverageJSON   json.RawMessage
	SubnetsJSON           json.RawMessage
	SurfacesJSON          json.RawMessage
	NetworkNamesJSON      json.RawMessage
	RiskFactorsJSON       json.RawMessage
	ActiveCoverageJSON    json.RawMessage
	DiscoveryMethodsJSON  json.RawMessage
	HostnamesJSON         json.RawMessage
	InternalIPsJSON       json.RawMessage
	InternalIPsV6JSON     json.RawMessage
	MACAddressesJSON      json.RawMessage
	GatewayIPsJSON        json.RawMessage
	GatewayMacsJSON       json.RawMessage
	TCPPortsJSON          json.RawMessage
	UDPPortsJSON          json.RawMessage
	RangerTagsJSON        json.RawMessage
	IDSecondaryJSON       json.RawMessage

	CollectedAt time.Time
}

// ConvertNetworkDiscovery converts an API network discovery device to NetworkDiscoveryData.
func ConvertNetworkDiscovery(d APINetworkDiscovery, collectedAt time.Time) *NetworkDiscoveryData {
	return &NetworkDiscoveryData{
		ResourceID:               d.ID,
		Name:                     d.Name,
		IPAddress:                d.IPAddress,
		Domain:                   d.Domain,
		SerialNumber:             d.SerialNumber,
		Category:                 d.Category,
		SubCategory:              d.SubCategory,
		ResourceType:             d.ResourceType,
		OS:                       d.OS,
		OSFamily:                 d.OSFamily,
		OSVersion:                d.OSVersion,
		OSNameVersion:            d.OSNameVersion,
		Architecture:             d.Architecture,
		Manufacturer:             d.Manufacturer,
		CPU:                      d.CPU,
		MemoryReadable:           d.MemoryReadable,
		NetworkName:              d.NetworkName,
		AssetStatus:              d.AssetStatus,
		AssetCriticality:         d.AssetCriticality,
		AssetEnvironment:         d.AssetEnvironment,
		InfectionStatus:          d.InfectionStatus,
		DeviceReview:             d.DeviceReview,
		EppUnsupportedUnknown:    d.EppUnsupportedUnknown,
		AssetContactEmail:        d.AssetContactEmail,
		LegacyIdentityPolicyName: d.LegacyIdentityPolicyName,
		PreviousOSType:           d.PreviousOSType,
		PreviousOSVersion:        d.PreviousOSVersion,
		PreviousDeviceFunction:   d.PreviousDeviceFunction,
		DetectedFromSite:         d.DetectedFromSite,
		S1AccountID:              d.S1AccountID,
		S1AccountName:            d.S1AccountName,
		S1SiteID:                 d.S1SiteID,
		S1SiteName:               d.S1SiteName,
		S1GroupID:                d.S1GroupID,
		S1GroupName:              d.S1GroupName,
		S1ScopeID:                d.S1ScopeID,
		S1ScopeLevel:             d.S1ScopeLevel,
		S1ScopePath:              d.S1ScopePath,
		S1OnboardedAccountName:   d.S1OnboardedAccountName,
		S1OnboardedGroupName:     d.S1OnboardedGroupName,
		S1OnboardedSiteName:      d.S1OnboardedSiteName,
		S1OnboardedScopeLevel:    d.S1OnboardedScopeLevel,
		S1OnboardedScopePath:     d.S1OnboardedScopePath,
		Memory:                   d.Memory,
		CoreCount:                d.CoreCount,
		S1ManagementID:           d.S1ManagementID,
		S1ScopeType:              d.S1ScopeType,
		S1OnboardedAccountID:     d.S1OnboardedAccountID,
		S1OnboardedGroupID:       d.S1OnboardedGroupID,
		S1OnboardedScopeID:       d.S1OnboardedScopeID,
		S1OnboardedSiteID:        d.S1OnboardedSiteID,
		IsAdConnector:            d.IsAdConnector,
		IsDcServer:               d.IsDcServer,
		AdsEnabled:               d.AdsEnabled,
		FirstSeenDt:              d.FirstSeenDt,
		LastUpdateDt:             d.LastUpdateDt,
		LastActiveDt:             d.LastActiveDt,
		LastRebootDt:             d.LastRebootDt,
		S1UpdatedAt:              d.S1UpdatedAt,
		AgentJSON:                d.AgentJSON,
		NetworkInterfacesJSON:    d.NetworkInterfacesJSON,
		AlertsJSON:               d.AlertsJSON,
		AlertsCountJSON:          d.AlertsCountJSON,
		DeviceReviewLogJSON:      d.DeviceReviewLogJSON,
		IdentityJSON:             d.IdentityJSON,
		NotesJSON:                d.NotesJSON,
		TagsJSON:                 d.TagsJSON,
		MissingCoverageJSON:      d.MissingCoverageJSON,
		SubnetsJSON:              d.SubnetsJSON,
		SurfacesJSON:             d.SurfacesJSON,
		NetworkNamesJSON:         d.NetworkNamesJSON,
		RiskFactorsJSON:          d.RiskFactorsJSON,
		ActiveCoverageJSON:       d.ActiveCoverageJSON,
		DiscoveryMethodsJSON:     d.DiscoveryMethodsJSON,
		HostnamesJSON:            d.HostnamesJSON,
		InternalIPsJSON:          d.InternalIPsJSON,
		InternalIPsV6JSON:        d.InternalIPsV6JSON,
		MACAddressesJSON:         d.MACAddressesJSON,
		GatewayIPsJSON:           d.GatewayIPsJSON,
		GatewayMacsJSON:          d.GatewayMacsJSON,
		TCPPortsJSON:             d.TCPPortsJSON,
		UDPPortsJSON:             d.UDPPortsJSON,
		RangerTagsJSON:           d.RangerTagsJSON,
		IDSecondaryJSON:          d.IDSecondaryJSON,
		CollectedAt:              collectedAt,
	}
}
