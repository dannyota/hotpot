package glbpackage

import (
	"encoding/json"
	"fmt"
	"time"

	glbv1 "danny.vn/gnode/services/glb/v1"
)

// GLBPackageData represents a converted global package ready for Ent insertion.
type GLBPackageData struct {
	ID                          string
	Name                        string
	Description                 string
	DescriptionEn               string
	DetailJSON                  json.RawMessage
	Enabled                     bool
	BaseSku                     string
	BaseConnectionRate          int
	BaseDomesticTrafficTotal    int
	BaseNonDomesticTrafficTotal int
	ConnectionSku               string
	DomesticTrafficSku          string
	NonDomesticTrafficSku       string
	CreatedAtAPI                string
	UpdatedAtAPI                string
	VlbPackagesJSON             json.RawMessage
	ProjectID                   string
	CollectedAt                 time.Time
}

// ConvertGLBPackage converts a GreenNode SDK GlobalPackage to GLBPackageData.
func ConvertGLBPackage(pkg *glbv1.GlobalPackage, projectID string, collectedAt time.Time) (*GLBPackageData, error) {
	data := &GLBPackageData{
		ID:                          pkg.ID,
		Name:                        pkg.Name,
		Description:                 pkg.Description,
		DescriptionEn:               pkg.DescriptionEn,
		Enabled:                     pkg.Enabled,
		BaseSku:                     pkg.BaseSku,
		BaseConnectionRate:          pkg.BaseConnectionRate,
		BaseDomesticTrafficTotal:    pkg.BaseDomesticTrafficTotal,
		BaseNonDomesticTrafficTotal: pkg.BaseNonDomesticTrafficTotal,
		ConnectionSku:               pkg.ConnectionSku,
		DomesticTrafficSku:          pkg.DomesticTrafficSku,
		NonDomesticTrafficSku:       pkg.NonDomesticTrafficSku,
		CreatedAtAPI:                pkg.CreatedAt,
		UpdatedAtAPI:                pkg.UpdatedAt,
		ProjectID:                   projectID,
		CollectedAt:                 collectedAt,
	}

	if pkg.Detail != nil {
		detailJSON, err := json.Marshal(pkg.Detail)
		if err != nil {
			return nil, fmt.Errorf("marshal detail for package %s: %w", pkg.ID, err)
		}
		data.DetailJSON = detailJSON
	}

	if len(pkg.VlbPackages) > 0 {
		vlbJSON, err := json.Marshal(pkg.VlbPackages)
		if err != nil {
			return nil, fmt.Errorf("marshal vlb packages for package %s: %w", pkg.ID, err)
		}
		data.VlbPackagesJSON = vlbJSON
	}

	return data, nil
}
