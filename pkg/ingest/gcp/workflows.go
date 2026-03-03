package gcp

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/dannyota/hotpot/pkg/base/temporalerr"
	"github.com/dannyota/hotpot/pkg/ingest"
)

// GCPInventoryWorkflowParams contains parameters for the GCP inventory workflow.
type GCPInventoryWorkflowParams struct{}

// GCPInventoryWorkflowResult contains the result of the GCP inventory workflow.
type GCPInventoryWorkflowResult struct {
	ProjectResults []ProjectResult

	// Compute
	TotalInstances        int
	TotalInterconnects    int
	TotalPacketMirrorings int
	TotalProjectMetadata  int

	// Container (GKE)
	TotalClusters int

	// IAM
	TotalServiceAccounts int

	// VPC Access
	TotalConnectors int

	// Storage
	TotalBuckets           int
	TotalBucketIamPolicies int

	// KMS
	TotalKeyRings   int
	TotalCryptoKeys int

	// Logging
	TotalSinks         int
	TotalLogBuckets    int
	TotalLogMetrics    int
	TotalLogExclusions int

	// DNS
	TotalManagedZones int
	TotalDNSPolicies  int

	// Secret Manager
	TotalSecrets int

	// Cloud SQL
	TotalSQLInstances int

	// Security Command Center (org-scoped)
	TotalSources  int
	TotalFindings int

	// Organization Policy (org-scoped)
	TotalConstraints int
	TotalOrgPolicies int

	// Service Usage
	TotalEnabledServices int

	// Cloud Functions
	TotalFunctions int

	// Memorystore Redis
	TotalRedisInstances int

	// Dataproc
	TotalDataprocClusters int

	// IAP
	TotalIAPSettings int
	TotalIAPPolicies int

	// AlloyDB
	TotalAlloyDBClusters int

	// Filestore
	TotalFilestoreInstances int

	// Pub/Sub
	TotalTopics        int
	TotalSubscriptions int

	// App Engine
	TotalApplications int
	TotalAppServices  int

	// Cloud Asset (org-scoped)
	TotalAssets         int
	TotalAssetPolicies  int
	TotalAssetResources int

	// Binary Authorization
	TotalBinAuthPolicies int
	TotalAttestors       int

	// Monitoring
	TotalAlertPolicies int
	TotalUptimeChecks  int

	// Cloud Run
	TotalRunServices  int
	TotalRunRevisions int

	// Access Context Manager (org-scoped)
	TotalAccessPolicies    int
	TotalAccessLevels      int
	TotalServicePerimeters int

	// Container Analysis
	TotalNotes       int
	TotalOccurrences int

	// Spanner
	TotalSpannerInstances int
	TotalSpannerDatabases int

	// BigQuery
	TotalDatasets int
	TotalTables   int

	// Bigtable
	TotalBigtableInstances int
	TotalBigtableClusters  int

	// Resource Manager (org-scoped)
	TotalProjects           int
	TotalOrganizations      int
	TotalFolders            int
	TotalOrgIamPolicies     int
	TotalFolderIamPolicies  int
	TotalProjectIamPolicies int
}

// ProjectResult contains the ingestion result for a single project.
type ProjectResult struct {
	ProjectID       string
	Error           string
	SkippedServices int

	// Compute
	InstanceCount        int
	InterconnectCount    int
	PacketMirroringCount int
	ProjectMetadataCount int

	// Container (GKE)
	ClusterCount int

	// IAM
	ServiceAccountCount int

	// VPC Access
	ConnectorCount int

	// Storage
	BucketCount          int
	BucketIamPolicyCount int

	// KMS
	KeyRingCount   int
	CryptoKeyCount int

	// Logging
	SinkCount         int
	LogBucketCount    int
	LogMetricCount    int
	LogExclusionCount int

	// DNS
	ManagedZoneCount int
	DNSPolicyCount   int

	// Secret Manager
	SecretCount int

	// Cloud SQL
	SQLInstanceCount int

	// Service Usage
	EnabledServiceCount int

	// Cloud Functions
	FunctionCount int

	// Memorystore Redis
	RedisInstanceCount int

	// Dataproc
	DataprocClusterCount int

	// IAP
	IAPSettingsCount int
	IAPPolicyCount   int

	// AlloyDB
	AlloyDBClusterCount int

	// Filestore
	FilestoreInstanceCount int

	// Pub/Sub
	TopicCount        int
	SubscriptionCount int

	// App Engine
	ApplicationCount int
	AppServiceCount  int

	// Binary Authorization
	BinAuthPolicyCount int
	AttestorCount      int

	// Monitoring
	AlertPolicyCount int
	UptimeCheckCount int

	// Cloud Run
	RunServiceCount  int
	RunRevisionCount int

	// Container Analysis
	NoteCount       int
	OccurrenceCount int

	// Spanner
	SpannerInstanceCount int
	SpannerDatabaseCount int

	// BigQuery
	DatasetCount int
	TableCount   int

	// Bigtable
	BigtableInstanceCount int
	BigtableClusterCount  int
}

// aggregateFunc is the function signature for merging a service result into the provider-level results.
type aggregateFunc = func(*GCPInventoryWorkflowResult, *ProjectResult, any)

// svcCall tracks a launched child workflow for later result collection.
type svcCall struct {
	projectID string
	svc       ingest.ServiceRegistration
	future    workflow.ChildWorkflowFuture
	result    any
}

// GCPInventoryWorkflow ingests all GCP resources across multiple projects.
// It orchestrates compute, GKE, IAM, and other GCP resource ingestion.
func GCPInventoryWorkflow(ctx workflow.Context, _ GCPInventoryWorkflowParams) (*GCPInventoryWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting GCPInventoryWorkflow")

	// Discover projects
	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	var discoverResult DiscoverProjectsResult
	err := workflow.ExecuteActivity(activityCtx, DiscoverProjectsActivity, DiscoverProjectsParams{}).
		Get(ctx, &discoverResult)
	if err != nil {
		logger.Error("Failed to discover projects", "error", err)
		return nil, temporalerr.PropagateNonRetryable(err)
	}

	logger.Info("Discovered projects", "count", len(discoverResult.ProjectIDs))

	// Child workflow options
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

	result := &GCPInventoryWorkflowResult{
		ProjectResults: make([]ProjectResult, 0, len(discoverResult.ProjectIDs)),
	}

	services := ingest.Services("gcp")

	// Phase 1: Discover enabled APIs for all projects in parallel
	apiFutures := make(map[string]workflow.Future, len(discoverResult.ProjectIDs))
	for _, pid := range discoverResult.ProjectIDs {
		apiFutures[pid] = workflow.ExecuteActivity(activityCtx, DiscoverEnabledAPIsActivity,
			DiscoverEnabledAPIsParams{ProjectID: pid})
	}

	enabledAPIs := make(map[string]map[string]bool, len(discoverResult.ProjectIDs))
	for _, pid := range discoverResult.ProjectIDs {
		var res DiscoverEnabledAPIsResult
		if err := apiFutures[pid].Get(ctx, &res); err != nil {
			logger.Error("Failed to discover enabled APIs; running all services",
				"projectID", pid, "error", err)
		} else {
			m := make(map[string]bool, len(res.EnabledAPIs))
			for _, api := range res.EnabledAPIs {
				m[api] = true
			}
			enabledAPIs[pid] = m
		}
	}

	// Phase 2: Launch all (project x service) combinations + global services in parallel
	var calls []svcCall
	skippedCount := make(map[string]int, len(discoverResult.ProjectIDs))

	// Regional services (per project)
	for _, pid := range discoverResult.ProjectIDs {
		apis := enabledAPIs[pid]
		for _, svc := range services {
			if svc.Scope != ingest.ScopeRegional {
				continue
			}
			if apis != nil && svc.APIName != "" && !apis[svc.APIName] {
				logger.Info("Skipping service: API not enabled",
					"service", svc.Name, "api", svc.APIName, "projectID", pid)
				skippedCount[pid]++
				continue
			}
			res := svc.NewResult()
			f := workflow.ExecuteChildWorkflow(ctx, svc.Workflow, svc.NewParams(pid, ""))
			calls = append(calls, svcCall{projectID: pid, svc: svc, future: f, result: res})
		}
	}

	// Global services (org-scoped, run once)
	for _, svc := range services {
		if svc.Scope != ingest.ScopeGlobal {
			continue
		}
		res := svc.NewResult()
		f := workflow.ExecuteChildWorkflow(ctx, svc.Workflow, svc.NewParams("", ""))
		calls = append(calls, svcCall{projectID: "", svc: svc, future: f, result: res})
	}

	// Phase 3: Collect all results and aggregate
	projectResults := make(map[string]*ProjectResult, len(discoverResult.ProjectIDs))
	for _, pid := range discoverResult.ProjectIDs {
		pr := &ProjectResult{ProjectID: pid, SkippedServices: skippedCount[pid]}
		projectResults[pid] = pr
	}

	for _, c := range calls {
		if err := c.future.Get(ctx, c.result); err != nil {
			if c.projectID != "" {
				logger.Error("Failed ingestion", "service", c.svc.Name, "projectID", c.projectID, "error", err)
				appendError(projectResults[c.projectID], err)
			} else {
				logger.Error("Failed ingestion", "service", c.svc.Name, "error", err)
			}
		} else {
			var pr *ProjectResult
			if c.projectID != "" {
				pr = projectResults[c.projectID]
			}
			c.svc.Aggregate.(aggregateFunc)(result, pr, c.result)
		}
	}

	// Build final project results slice (preserving original order)
	for _, pid := range discoverResult.ProjectIDs {
		result.ProjectResults = append(result.ProjectResults, *projectResults[pid])
	}

	logger.Info("Completed GCPInventoryWorkflow",
		"projectCount", len(discoverResult.ProjectIDs),
		"totalInstances", result.TotalInstances,
		"totalClusters", result.TotalClusters,
		"totalServiceAccounts", result.TotalServiceAccounts,
		"totalConnectors", result.TotalConnectors,
		"totalBuckets", result.TotalBuckets,
		"totalKeyRings", result.TotalKeyRings,
		"totalCryptoKeys", result.TotalCryptoKeys,
		"totalSinks", result.TotalSinks,
		"totalLogBuckets", result.TotalLogBuckets,
		"totalManagedZones", result.TotalManagedZones,
		"totalSecrets", result.TotalSecrets,
		"totalSQLInstances", result.TotalSQLInstances,
		"totalLogMetrics", result.TotalLogMetrics,
		"totalLogExclusions", result.TotalLogExclusions,
		"totalDNSPolicies", result.TotalDNSPolicies,
		"totalBucketIamPolicies", result.TotalBucketIamPolicies,
		"totalInterconnects", result.TotalInterconnects,
		"totalPacketMirrorings", result.TotalPacketMirrorings,
		"totalProjectMetadata", result.TotalProjectMetadata,
		"totalSources", result.TotalSources,
		"totalFindings", result.TotalFindings,
		"totalConstraints", result.TotalConstraints,
		"totalOrgPolicies", result.TotalOrgPolicies,
		"totalEnabledServices", result.TotalEnabledServices,
		"totalFunctions", result.TotalFunctions,
		"totalRedisInstances", result.TotalRedisInstances,
		"totalDataprocClusters", result.TotalDataprocClusters,
		"totalIAPSettings", result.TotalIAPSettings,
		"totalIAPPolicies", result.TotalIAPPolicies,
		"totalAlloyDBClusters", result.TotalAlloyDBClusters,
		"totalFilestoreInstances", result.TotalFilestoreInstances,
		"totalTopics", result.TotalTopics,
		"totalSubscriptions", result.TotalSubscriptions,
		"totalApplications", result.TotalApplications,
		"totalAppServices", result.TotalAppServices,
		"totalAssets", result.TotalAssets,
		"totalAssetPolicies", result.TotalAssetPolicies,
		"totalAssetResources", result.TotalAssetResources,
		"totalBinAuthPolicies", result.TotalBinAuthPolicies,
		"totalAttestors", result.TotalAttestors,
		"totalAlertPolicies", result.TotalAlertPolicies,
		"totalUptimeChecks", result.TotalUptimeChecks,
		"totalRunServices", result.TotalRunServices,
		"totalRunRevisions", result.TotalRunRevisions,
		"totalAccessPolicies", result.TotalAccessPolicies,
		"totalAccessLevels", result.TotalAccessLevels,
		"totalServicePerimeters", result.TotalServicePerimeters,
		"totalNotes", result.TotalNotes,
		"totalOccurrences", result.TotalOccurrences,
		"totalSpannerInstances", result.TotalSpannerInstances,
		"totalSpannerDatabases", result.TotalSpannerDatabases,
		"totalDatasets", result.TotalDatasets,
		"totalTables", result.TotalTables,
		"totalBigtableInstances", result.TotalBigtableInstances,
		"totalBigtableClusters", result.TotalBigtableClusters,
		"totalProjects", result.TotalProjects,
		"totalOrganizations", result.TotalOrganizations,
		"totalFolders", result.TotalFolders,
		"totalOrgIamPolicies", result.TotalOrgIamPolicies,
		"totalFolderIamPolicies", result.TotalFolderIamPolicies,
		"totalProjectIamPolicies", result.TotalProjectIamPolicies,
	)

	return result, nil
}

func appendError(pr *ProjectResult, err error) {
	if pr.Error == "" {
		pr.Error = err.Error()
	} else {
		pr.Error += "; " + err.Error()
	}
}
