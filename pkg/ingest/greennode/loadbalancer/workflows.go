package loadbalancer

import (
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/ingest/greennode/loadbalancer/certificate"
	"danny.vn/hotpot/pkg/ingest/greennode/loadbalancer/lb"
	"danny.vn/hotpot/pkg/ingest/greennode/loadbalancer/lbpackage"
)

// GreenNodeLoadBalancerWorkflowParams contains parameters for the load balancer workflow.
type GreenNodeLoadBalancerWorkflowParams struct {
	ProjectID string
	Region    string
}

// GreenNodeLoadBalancerWorkflowResult contains the result of the load balancer workflow.
type GreenNodeLoadBalancerWorkflowResult struct {
	LBCount          int
	CertificateCount int
	PackageCount     int
}

// GreenNodeLoadBalancerWorkflow orchestrates GreenNode load balancer ingestion.
func GreenNodeLoadBalancerWorkflow(ctx workflow.Context, params GreenNodeLoadBalancerWorkflowParams) (*GreenNodeLoadBalancerWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GreenNodeLoadBalancerWorkflow", "projectID", params.ProjectID, "region", params.Region)

	result := &GreenNodeLoadBalancerWorkflowResult{}

	childOpts := workflow.ChildWorkflowOptions{}
	childCtx := workflow.WithChildOptions(ctx, childOpts)

	// Load Balancers
	var lbResult lb.GreenNodeLoadBalancerLBWorkflowResult
	err := workflow.ExecuteChildWorkflow(childCtx, lb.GreenNodeLoadBalancerLBWorkflow, lb.GreenNodeLoadBalancerLBWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &lbResult)
	if err != nil {
		logger.Error("Failed to ingest load balancers", "error", err)
	} else {
		result.LBCount = lbResult.LBCount
	}

	// Certificates
	var certResult certificate.GreenNodeLoadBalancerCertificateWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, certificate.GreenNodeLoadBalancerCertificateWorkflow, certificate.GreenNodeLoadBalancerCertificateWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &certResult)
	if err != nil {
		logger.Error("Failed to ingest certificates", "error", err)
	} else {
		result.CertificateCount = certResult.CertificateCount
	}

	// Packages
	var pkgResult lbpackage.GreenNodeLoadBalancerPackageWorkflowResult
	err = workflow.ExecuteChildWorkflow(childCtx, lbpackage.GreenNodeLoadBalancerPackageWorkflow, lbpackage.GreenNodeLoadBalancerPackageWorkflowParams{
		ProjectID: params.ProjectID,
		Region:    params.Region,
	}).Get(ctx, &pkgResult)
	if err != nil {
		logger.Error("Failed to ingest packages", "error", err)
	} else {
		result.PackageCount = pkgResult.PackageCount
	}

	logger.Info("Completed GreenNodeLoadBalancerWorkflow",
		"lbCount", result.LBCount,
		"certificateCount", result.CertificateCount,
		"packageCount", result.PackageCount,
	)

	return result, nil
}
