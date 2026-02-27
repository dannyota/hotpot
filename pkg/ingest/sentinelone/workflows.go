package sentinelone

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/account"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/agent"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/app_inventory"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/endpoint_app"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/group"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/network_discovery"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/ranger_device"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/ranger_gateway"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/ranger_setting"
	"github.com/dannyota/hotpot/pkg/ingest/sentinelone/site"
)

// S1InventoryWorkflowResult contains the result of SentinelOne inventory collection.
type S1InventoryWorkflowResult struct {
	AccountCount       int
	AgentCount         int
	GroupCount         int
	SiteCount          int
	RangerDeviceCount  int
	RangerGatewayCount int
	RangerSettingCount    int
	NetworkDiscoveryCount int
	AppInventoryCount     int
	EndpointAppCount   int
}

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

	// Execute account workflow
	var accountResult account.S1AccountWorkflowResult
	err := workflow.ExecuteChildWorkflow(ctx, account.S1AccountWorkflow).Get(ctx, &accountResult)
	if err != nil {
		logger.Error("Failed to execute S1AccountWorkflow", "error", err)
	} else {
		result.AccountCount = accountResult.AccountCount
	}

	// Execute agent workflow
	var agentResult agent.S1AgentWorkflowResult
	err = workflow.ExecuteChildWorkflow(ctx, agent.S1AgentWorkflow).Get(ctx, &agentResult)
	if err != nil {
		logger.Error("Failed to execute S1AgentWorkflow", "error", err)
	} else {
		result.AgentCount = agentResult.AgentCount
	}

	// Execute site workflow
	var siteResult site.S1SiteWorkflowResult
	err = workflow.ExecuteChildWorkflow(ctx, site.S1SiteWorkflow).Get(ctx, &siteResult)
	if err != nil {
		logger.Error("Failed to execute S1SiteWorkflow", "error", err)
	} else {
		result.SiteCount = siteResult.SiteCount
	}

	// Execute group workflow
	var groupResult group.S1GroupWorkflowResult
	err = workflow.ExecuteChildWorkflow(ctx, group.S1GroupWorkflow).Get(ctx, &groupResult)
	if err != nil {
		logger.Error("Failed to execute S1GroupWorkflow", "error", err)
	} else {
		result.GroupCount = groupResult.GroupCount
	}

	// Execute ranger device workflow
	var rangerDeviceResult ranger_device.S1RangerDeviceWorkflowResult
	err = workflow.ExecuteChildWorkflow(ctx, ranger_device.S1RangerDeviceWorkflow).Get(ctx, &rangerDeviceResult)
	if err != nil {
		logger.Error("Failed to execute S1RangerDeviceWorkflow", "error", err)
	} else {
		result.RangerDeviceCount = rangerDeviceResult.DeviceCount
	}

	// Execute ranger gateway workflow
	var rangerGatewayResult ranger_gateway.S1RangerGatewayWorkflowResult
	err = workflow.ExecuteChildWorkflow(ctx, ranger_gateway.S1RangerGatewayWorkflow).Get(ctx, &rangerGatewayResult)
	if err != nil {
		logger.Error("Failed to execute S1RangerGatewayWorkflow", "error", err)
	} else {
		result.RangerGatewayCount = rangerGatewayResult.GatewayCount
	}

	// Execute ranger setting workflow
	var rangerSettingResult ranger_setting.S1RangerSettingWorkflowResult
	err = workflow.ExecuteChildWorkflow(ctx, ranger_setting.S1RangerSettingWorkflow).Get(ctx, &rangerSettingResult)
	if err != nil {
		logger.Error("Failed to execute S1RangerSettingWorkflow", "error", err)
	} else {
		result.RangerSettingCount = rangerSettingResult.SettingCount
	}

	// Execute network discovery workflow
	var networkDiscoveryResult network_discovery.S1NetworkDiscoveryWorkflowResult
	err = workflow.ExecuteChildWorkflow(ctx, network_discovery.S1NetworkDiscoveryWorkflow).Get(ctx, &networkDiscoveryResult)
	if err != nil {
		logger.Error("Failed to execute S1NetworkDiscoveryWorkflow", "error", err)
	} else {
		result.NetworkDiscoveryCount = networkDiscoveryResult.DeviceCount
	}

	// Execute app inventory workflow
	var appInventoryResult app_inventory.S1AppInventoryWorkflowResult
	err = workflow.ExecuteChildWorkflow(ctx, app_inventory.S1AppInventoryWorkflow).Get(ctx, &appInventoryResult)
	if err != nil {
		logger.Error("Failed to execute S1AppInventoryWorkflow", "error", err)
	} else {
		result.AppInventoryCount = appInventoryResult.AppCount
	}

	// Execute endpoint app workflow
	var endpointAppResult endpoint_app.S1EndpointAppWorkflowResult
	err = workflow.ExecuteChildWorkflow(ctx, endpoint_app.S1EndpointAppWorkflow).Get(ctx, &endpointAppResult)
	if err != nil {
		logger.Error("Failed to execute S1EndpointAppWorkflow", "error", err)
	} else {
		result.EndpointAppCount = endpointAppResult.AppCount
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
