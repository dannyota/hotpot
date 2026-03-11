package targethttpsproxy

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// TargetHttpsProxyData holds converted target HTTPS proxy data ready for Ent insertion.
type TargetHttpsProxyData struct {
	ID                      string
	Name                    string
	Description             string
	CreationTimestamp       string
	SelfLink                string
	Fingerprint             string
	UrlMap                  string
	QuicOverride            string
	ServerTlsPolicy         string
	AuthorizationPolicy     string
	CertificateMap          string
	SslPolicy               string
	TlsEarlyData            string
	ProxyBind               bool
	HttpKeepAliveTimeoutSec int32
	SslCertificatesJSON     []interface{}
	Region                  string
	ProjectID               string
	CollectedAt             time.Time
}

// ConvertTargetHttpsProxy converts a GCP API TargetHttpsProxy to Ent-compatible data.
func ConvertTargetHttpsProxy(thsp *computepb.TargetHttpsProxy, projectID string, collectedAt time.Time) (*TargetHttpsProxyData, error) {
	data := &TargetHttpsProxyData{
		ID:                      fmt.Sprintf("%d", thsp.GetId()),
		Name:                    thsp.GetName(),
		Description:             thsp.GetDescription(),
		CreationTimestamp:       thsp.GetCreationTimestamp(),
		SelfLink:                thsp.GetSelfLink(),
		Fingerprint:             thsp.GetFingerprint(),
		UrlMap:                  thsp.GetUrlMap(),
		QuicOverride:            thsp.GetQuicOverride(),
		ServerTlsPolicy:         thsp.GetServerTlsPolicy(),
		AuthorizationPolicy:     thsp.GetAuthorizationPolicy(),
		CertificateMap:          thsp.GetCertificateMap(),
		SslPolicy:               thsp.GetSslPolicy(),
		TlsEarlyData:            thsp.GetTlsEarlyData(),
		ProxyBind:               thsp.GetProxyBind(),
		HttpKeepAliveTimeoutSec: thsp.GetHttpKeepAliveTimeoutSec(),
		Region:                  thsp.GetRegion(),
		ProjectID:               projectID,
		CollectedAt:             collectedAt,
	}

	if thsp.SslCertificates != nil {
		certBytes, err := json.Marshal(thsp.SslCertificates)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ssl certificates for target HTTPS proxy %s: %w", thsp.GetName(), err)
		}
		if err := json.Unmarshal(certBytes, &data.SslCertificatesJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal ssl certificates: %w", err)
		}
	}

	return data, nil
}
