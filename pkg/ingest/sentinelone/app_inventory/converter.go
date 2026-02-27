package app_inventory

import "time"

// AppInventoryData holds converted app inventory data ready for Ent insertion.
type AppInventoryData struct {
	ResourceID               string
	ApplicationName          string
	ApplicationVendor        string
	EndpointsCount           int
	ApplicationVersionsCount int
	Estimate                 bool
	CollectedAt              time.Time
}

// ConvertAppInventory converts an API app inventory entry to AppInventoryData.
func ConvertAppInventory(app APIAppInventory, collectedAt time.Time) *AppInventoryData {
	return &AppInventoryData{
		ResourceID:               app.ApplicationName + "||" + app.ApplicationVendor,
		ApplicationName:          app.ApplicationName,
		ApplicationVendor:        app.ApplicationVendor,
		EndpointsCount:           app.EndpointsCount,
		ApplicationVersionsCount: app.ApplicationVersionsCount,
		Estimate:                 app.Estimate,
		CollectedAt:              collectedAt,
	}
}
