package targethttpproxy

import (
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// TargetHttpProxyData holds converted target HTTP proxy data ready for Ent insertion.
type TargetHttpProxyData struct {
	ID                      string
	Name                    string
	Description             string
	CreationTimestamp       string
	SelfLink                string
	Fingerprint             string
	UrlMap                  string
	ProxyBind               bool
	HttpKeepAliveTimeoutSec int32
	Region                  string
	ProjectID               string
	CollectedAt             time.Time
}

// ConvertTargetHttpProxy converts a GCP API TargetHttpProxy to Ent-compatible data.
func ConvertTargetHttpProxy(thp *computepb.TargetHttpProxy, projectID string, collectedAt time.Time) *TargetHttpProxyData {
	return &TargetHttpProxyData{
		ID:                      fmt.Sprintf("%d", thp.GetId()),
		Name:                    thp.GetName(),
		Description:             thp.GetDescription(),
		CreationTimestamp:       thp.GetCreationTimestamp(),
		SelfLink:                thp.GetSelfLink(),
		Fingerprint:             thp.GetFingerprint(),
		UrlMap:                  thp.GetUrlMap(),
		ProxyBind:               thp.GetProxyBind(),
		HttpKeepAliveTimeoutSec: thp.GetHttpKeepAliveTimeoutSec(),
		Region:                  thp.GetRegion(),
		ProjectID:               projectID,
		CollectedAt:             collectedAt,
	}
}
