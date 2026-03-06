package sentinelone

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"danny.vn/hotpot/pkg/ingest"
)

// S1InventoryWorkflowResult contains the result of SentinelOne inventory collection.
type S1InventoryWorkflowResult struct {
	AccountCount          int
	AgentCount            int
	GroupCount            int
	SiteCount             int
	RangerDeviceCount     int
	RangerGatewayCount    int
	RangerSettingCount    int
	NetworkDiscoveryCount int
	AppInventoryCount     int
	EndpointAppCount      int
}

// aggregateFunc is the function signature for merging a service result into the provider result.
type aggregateFunc = func(*S1InventoryWorkflowResult, any)

// S1InventoryWorkflow orchestrates SentinelOne inventory collection.
func S1InventoryWorkflow(ctx workflow.Context) (*S1InventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting S1InventoryWorkflow")

	childOpts := workflow.ChildWorkflowOptions{
		WorkflowExecutionTimeout: 60 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithChildOptions(ctx, childOpts)

	result := &S1InventoryWorkflowResult{}

	for _, svc := range ingest.Services("sentinelone") {
		res := svc.NewResult()
		err := workflow.ExecuteChildWorkflow(ctx, svc.Workflow).Get(ctx, res)
		if err != nil {
			logger.Error("Failed ingestion", "service", svc.Name, "error", err)
		} else {
			svc.Aggregate.(aggregateFunc)(result, res)
		}
	}

	logger.Info("Completed S1InventoryWorkflow",
		"accounts", result.AccountCount,
		"agents", result.AgentCount,
		"groups", result.GroupCount,
		"sites", result.SiteCount,
		"rangerDevices", result.RangerDeviceCount,
		"rangerGateways", result.RangerGatewayCount,
		"rangerSettings", result.RangerSettingCount,
		"networkDiscoveries", result.NetworkDiscoveryCount,
		"appInventory", result.AppInventoryCount,
		"endpointApps", result.EndpointAppCount,
	)

	return result, nil
}
