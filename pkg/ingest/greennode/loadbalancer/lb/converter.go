package lb

import (
	"encoding/json"
	"fmt"
	"time"

	lbv2 "danny.vn/greennode/services/loadbalancer/v2"
)

// LBData represents a converted load balancer ready for Ent insertion.
type LBData struct {
	ID                 string
	Name               string
	DisplayStatus      string
	Address            string
	PrivateSubnetID    string
	PrivateSubnetCidr  string
	Type               string
	DisplayType        string
	LoadBalancerSchema string
	PackageID          string
	Description        string
	Location           string
	CreatedAtAPI       string
	UpdatedAtAPI       string
	ProgressStatus     string
	Status             string
	BackendSubnetID    string
	Internal           bool
	AutoScalable       bool
	ZoneID             string
	MinSize            int
	MaxSize            int
	TotalNodes         int
	NodesJSON          json.RawMessage
	Region             string
	ProjectID          string
	CollectedAt        time.Time

	Listeners []ListenerData
	Pools     []PoolData
}

// ListenerData represents a converted listener ready for Ent insertion.
type ListenerData struct {
	ListenerID                      string
	Name                            string
	Description                     string
	Protocol                        string
	ProtocolPort                    int
	ConnectionLimit                 int
	DefaultPoolID                   string
	DefaultPoolName                 string
	TimeoutClient                   int
	TimeoutMember                   int
	TimeoutConnection               int
	AllowedCidrs                    string
	CertificateAuthoritiesJSON      json.RawMessage
	DisplayStatus                   string
	CreatedAtAPI                    string
	UpdatedAtAPI                    string
	DefaultCertificateAuthority     *string
	ClientCertificateAuthentication *string
	ProgressStatus                  string
	InsertHeadersJSON               json.RawMessage
	PoliciesJSON                    json.RawMessage
}

// PoolData represents a converted pool ready for Ent insertion.
type PoolData struct {
	PoolID            string
	Name              string
	Protocol          string
	Description       string
	LoadBalanceMethod string
	Status            string
	Stickiness        bool
	TLSEncryption     bool
	MembersJSON       json.RawMessage
	HealthMonitorJSON json.RawMessage
}

// ConvertLB converts a GreenNode SDK LoadBalancer to LBData.
func ConvertLB(lb *lbv2.LoadBalancer, listeners []*lbv2.Listener, pools []*lbv2.Pool, projectID, region string, collectedAt time.Time) (*LBData, error) {
	data := &LBData{
		ID:                 lb.UUID,
		Name:               lb.Name,
		DisplayStatus:      lb.DisplayStatus,
		Address:            lb.Address,
		PrivateSubnetID:    lb.PrivateSubnetID,
		PrivateSubnetCidr:  lb.PrivateSubnetCidr,
		Type:               lb.Type,
		DisplayType:        lb.DisplayType,
		LoadBalancerSchema: lb.LoadBalancerSchema,
		PackageID:          lb.PackageID,
		Description:        lb.Description,
		Location:           lb.Location,
		CreatedAtAPI:       lb.CreatedAt,
		UpdatedAtAPI:       lb.UpdatedAt,
		ProgressStatus:     lb.ProgressStatus,
		Status:             lb.Status,
		BackendSubnetID:    lb.BackendSubnetID,
		Internal:           lb.Internal,
		AutoScalable:       lb.AutoScalable,
		ZoneID:             lb.ZoneID,
		MinSize:            lb.MinSize,
		MaxSize:            lb.MaxSize,
		TotalNodes:         lb.TotalNodes,
		Region:             region,
		ProjectID:          projectID,
		CollectedAt:        collectedAt,
	}

	// Marshal nodes to JSON
	if len(lb.Nodes) > 0 {
		nodesJSON, err := json.Marshal(lb.Nodes)
		if err != nil {
			return nil, fmt.Errorf("marshal nodes for LB %s: %w", lb.Name, err)
		}
		data.NodesJSON = nodesJSON
	}

	// Convert listeners
	data.Listeners = ConvertListeners(listeners)

	// Convert pools
	var err error
	data.Pools, err = ConvertPools(pools)
	if err != nil {
		return nil, fmt.Errorf("convert pools for LB %s: %w", lb.Name, err)
	}

	return data, nil
}

// ConvertListeners converts SDK listeners to ListenerData.
func ConvertListeners(listeners []*lbv2.Listener) []ListenerData {
	if len(listeners) == 0 {
		return nil
	}
	result := make([]ListenerData, 0, len(listeners))
	for _, l := range listeners {
		ld := ListenerData{
			ListenerID:                      l.UUID,
			Name:                            l.Name,
			Description:                     l.Description,
			Protocol:                        l.Protocol,
			ProtocolPort:                    l.ProtocolPort,
			ConnectionLimit:                 l.ConnectionLimit,
			DefaultPoolID:                   l.DefaultPoolID,
			DefaultPoolName:                 l.DefaultPoolName,
			TimeoutClient:                   l.TimeoutClient,
			TimeoutMember:                   l.TimeoutMember,
			TimeoutConnection:               l.TimeoutConnection,
			AllowedCidrs:                    l.AllowedCidrs,
			DisplayStatus:                   l.DisplayStatus,
			CreatedAtAPI:                    l.CreatedAt,
			UpdatedAtAPI:                    l.UpdatedAt,
			DefaultCertificateAuthority:     l.DefaultCertificateAuthority,
			ClientCertificateAuthentication: l.ClientCertificateAuthentication,
			ProgressStatus:                  l.ProgressStatus,
		}

		// Marshal certificate authorities to JSON
		if len(l.CertificateAuthorities) > 0 {
			caJSON, err := json.Marshal(l.CertificateAuthorities)
			if err == nil {
				ld.CertificateAuthoritiesJSON = caJSON
			}
		}

		// Marshal insert headers to JSON
		if len(l.InsertHeaders) > 0 {
			ihJSON, err := json.Marshal(l.InsertHeaders)
			if err == nil {
				ld.InsertHeadersJSON = ihJSON
			}
		}

		result = append(result, ld)
	}
	return result
}

// ConvertPools converts SDK pools to PoolData.
func ConvertPools(pools []*lbv2.Pool) ([]PoolData, error) {
	if len(pools) == 0 {
		return nil, nil
	}
	result := make([]PoolData, 0, len(pools))
	for _, p := range pools {
		pd := PoolData{
			PoolID:            p.UUID,
			Name:              p.Name,
			Protocol:          p.Protocol,
			Description:       p.Description,
			LoadBalanceMethod: p.LoadBalanceMethod,
			Status:            p.Status,
			Stickiness:        p.Stickiness,
			TLSEncryption:     p.TLSEncryption,
		}

		// Marshal members to JSON
		if p.Members != nil {
			membersJSON, err := json.Marshal(p.Members)
			if err != nil {
				return nil, fmt.Errorf("marshal members for pool %s: %w", p.Name, err)
			}
			pd.MembersJSON = membersJSON
		}

		// Marshal health monitor to JSON
		if p.HealthMonitor != nil {
			hmJSON, err := json.Marshal(p.HealthMonitor)
			if err != nil {
				return nil, fmt.Errorf("marshal health monitor for pool %s: %w", p.Name, err)
			}
			pd.HealthMonitorJSON = hmJSON
		}

		result = append(result, pd)
	}
	return result, nil
}
