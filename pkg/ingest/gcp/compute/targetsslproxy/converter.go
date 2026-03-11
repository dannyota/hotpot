package targetsslproxy

import (
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/compute/apiv1/computepb"
)

// TargetSslProxyData holds converted target SSL proxy data ready for Ent insertion.
type TargetSslProxyData struct {
	ID                  string
	Name                string
	Description         string
	CreationTimestamp   string
	SelfLink            string
	Service             string
	ProxyHeader         string
	CertificateMap      string
	SslPolicy           string
	SslCertificatesJSON []interface{}
	ProjectID           string
	CollectedAt         time.Time
}

// ConvertTargetSslProxy converts a GCP API TargetSslProxy to Ent-compatible data.
func ConvertTargetSslProxy(tsp *computepb.TargetSslProxy, projectID string, collectedAt time.Time) (*TargetSslProxyData, error) {
	data := &TargetSslProxyData{
		ID:                fmt.Sprintf("%d", tsp.GetId()),
		Name:              tsp.GetName(),
		Description:       tsp.GetDescription(),
		CreationTimestamp: tsp.GetCreationTimestamp(),
		SelfLink:          tsp.GetSelfLink(),
		Service:           tsp.GetService(),
		ProxyHeader:       tsp.GetProxyHeader(),
		CertificateMap:    tsp.GetCertificateMap(),
		SslPolicy:         tsp.GetSslPolicy(),
		ProjectID:         projectID,
		CollectedAt:       collectedAt,
	}

	if tsp.SslCertificates != nil {
		certBytes, err := json.Marshal(tsp.SslCertificates)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal ssl certificates for target SSL proxy %s: %w", tsp.GetName(), err)
		}
		if err := json.Unmarshal(certBytes, &data.SslCertificatesJSON); err != nil {
			return nil, fmt.Errorf("failed to unmarshal ssl certificates: %w", err)
		}
	}

	return data, nil
}
