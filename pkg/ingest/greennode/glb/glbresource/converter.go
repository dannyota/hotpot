package glbresource

import (
	"encoding/json"
	"fmt"
	"time"

	glbv1 "danny.vn/gnode/services/glb/v1"
)

// GLBData represents a converted global load balancer ready for Ent insertion.
type GLBData struct {
	ID           string
	Name         string
	Description  string
	Status       string
	Package      string
	Type         string
	UserID       int
	VipsJSON     json.RawMessage
	DomainsJSON  json.RawMessage
	CreatedAtAPI string
	UpdatedAtAPI string
	DeletedAtAPI string
	ProjectID    string
	CollectedAt  time.Time

	Listeners []GLBListenerData
	Pools     []GLBPoolData
}

// GLBListenerData represents a converted global listener.
type GLBListenerData struct {
	ListenerID        string
	Name              string
	Description       string
	Protocol          string
	Port              int
	GlobalPoolID      string
	TimeoutClient     int
	TimeoutMember     int
	TimeoutConnection int
	AllowedCidrs      string
	Headers           *string
	Status            string
	CreatedAtAPI      string
	UpdatedAtAPI      string
	DeletedAtAPI      *string
}

// GLBPoolData represents a converted global pool.
type GLBPoolData struct {
	PoolID          string
	Name            string
	Description     string
	Algorithm       string
	StickySession   *string
	TLSEnabled      *string
	Protocol        string
	Status          string
	HealthJSON      json.RawMessage
	PoolMembersJSON json.RawMessage
	CreatedAtAPI    string
	UpdatedAtAPI    string
	DeletedAtAPI    *string
}

// ConvertGLB converts a GreenNode SDK GlobalLoadBalancer to GLBData.
func ConvertGLB(glb *glbv1.GlobalLoadBalancer, projectID string, collectedAt time.Time) (*GLBData, error) {
	data := &GLBData{
		ID:           glb.ID,
		Name:         glb.Name,
		Description:  glb.Description,
		Status:       glb.Status,
		Package:      glb.Package,
		Type:         glb.Type,
		UserID:       glb.UserID,
		CreatedAtAPI: glb.CreatedAt,
		UpdatedAtAPI: glb.UpdatedAt,
		DeletedAtAPI: glb.DeletedAt,
		ProjectID:    projectID,
		CollectedAt:  collectedAt,
	}

	if len(glb.Vips) > 0 {
		vipsJSON, err := json.Marshal(glb.Vips)
		if err != nil {
			return nil, fmt.Errorf("marshal vips for GLB %s: %w", glb.ID, err)
		}
		data.VipsJSON = vipsJSON
	}

	if len(glb.Domains) > 0 {
		domainsJSON, err := json.Marshal(glb.Domains)
		if err != nil {
			return nil, fmt.Errorf("marshal domains for GLB %s: %w", glb.ID, err)
		}
		data.DomainsJSON = domainsJSON
	}

	return data, nil
}

// ConvertListeners converts SDK GlobalListener items to GLBListenerData.
func ConvertListeners(listeners []*glbv1.GlobalListener) []GLBListenerData {
	if len(listeners) == 0 {
		return nil
	}
	result := make([]GLBListenerData, 0, len(listeners))
	for _, l := range listeners {
		result = append(result, GLBListenerData{
			ListenerID:        l.ID,
			Name:              l.Name,
			Description:       l.Description,
			Protocol:          l.Protocol,
			Port:              l.Port,
			GlobalPoolID:      l.GlobalPoolID,
			TimeoutClient:     l.TimeoutClient,
			TimeoutMember:     l.TimeoutMember,
			TimeoutConnection: l.TimeoutConnection,
			AllowedCidrs:      l.AllowedCidrs,
			Headers:           l.Headers,
			Status:            l.Status,
			CreatedAtAPI:      l.CreatedAt,
			UpdatedAtAPI:      l.UpdatedAt,
			DeletedAtAPI:      l.DeletedAt,
		})
	}
	return result
}

// ConvertPools converts SDK GlobalPool items to GLBPoolData with pool members.
func ConvertPools(pools []*glbv1.GlobalPool, poolMembers map[string][]*glbv1.GlobalPoolMember) ([]GLBPoolData, error) {
	if len(pools) == 0 {
		return nil, nil
	}
	result := make([]GLBPoolData, 0, len(pools))
	for _, p := range pools {
		pd := GLBPoolData{
			PoolID:        p.ID,
			Name:          p.Name,
			Description:   p.Description,
			Algorithm:     p.Algorithm,
			StickySession: p.StickySession,
			TLSEnabled:    p.TLSEnabled,
			Protocol:      p.Protocol,
			Status:        p.Status,
			CreatedAtAPI:  p.CreatedAt,
			UpdatedAtAPI:  p.UpdatedAt,
			DeletedAtAPI:  p.DeletedAt,
		}

		if p.Health != nil {
			healthJSON, err := json.Marshal(p.Health)
			if err != nil {
				return nil, fmt.Errorf("marshal health for pool %s: %w", p.ID, err)
			}
			pd.HealthJSON = healthJSON
		}

		if members, ok := poolMembers[p.ID]; ok && len(members) > 0 {
			membersJSON, err := json.Marshal(members)
			if err != nil {
				return nil, fmt.Errorf("marshal pool members for pool %s: %w", p.ID, err)
			}
			pd.PoolMembersJSON = membersJSON
		}

		result = append(result, pd)
	}
	return result, nil
}
