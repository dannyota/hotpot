package application

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/appengine/apiv1/appenginepb"
)

// ApplicationData holds converted App Engine application data ready for Ent insertion.
type ApplicationData struct {
	ID                      string
	Name                    string
	AuthDomain              string
	LocationID              string
	CodeBucket              string
	DefaultCookieExpiration string
	ServingStatus           int32
	DefaultHostname         string
	DefaultBucket           string
	GcrDomain               string
	DatabaseType            int32
	FeatureSettingsJSON     json.RawMessage
	IapJSON                 json.RawMessage
	DispatchRulesJSON       json.RawMessage
	ProjectID               string
	CollectedAt             time.Time
}

// ConvertApplication converts a raw GCP API App Engine application to Ent-compatible data.
func ConvertApplication(app *appenginepb.Application, projectID string, collectedAt time.Time) (*ApplicationData, error) {
	data := &ApplicationData{
		ID:                      app.GetName(),
		Name:                    app.GetName(),
		AuthDomain:              app.GetAuthDomain(),
		LocationID:              app.GetLocationId(),
		CodeBucket:              app.GetCodeBucket(),
		DefaultCookieExpiration: app.GetDefaultCookieExpiration().String(),
		ServingStatus:           int32(app.GetServingStatus()),
		DefaultHostname:         app.GetDefaultHostname(),
		DefaultBucket:           app.GetDefaultBucket(),
		GcrDomain:               app.GetGcrDomain(),
		DatabaseType:            int32(app.GetDatabaseType()),
		ProjectID:               projectID,
		CollectedAt:             collectedAt,
	}

	var err error
	if app.FeatureSettings != nil {
		data.FeatureSettingsJSON, err = json.Marshal(app.FeatureSettings)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal feature settings for application %s: %w", app.GetName(), err)
		}
	}
	if app.Iap != nil {
		data.IapJSON, err = json.Marshal(app.Iap)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal IAP for application %s: %w", app.GetName(), err)
		}
	}
	if len(app.DispatchRules) > 0 {
		data.DispatchRulesJSON, err = json.Marshal(app.DispatchRules)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal dispatch rules for application %s: %w", app.GetName(), err)
		}
	}

	return data, nil
}
