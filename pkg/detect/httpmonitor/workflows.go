package httpmonitor

import (
	"errors"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// HttpMonitorAnomalyResult holds the combined result of the anomaly detection workflow.
type HttpMonitorAnomalyResult struct {
	RateResult           DetectRateAnomaliesResult
	ErrorResult          DetectErrorBurstsResult
	SuspiciousResult     DetectSuspiciousPatternsResult
	MethodMismatchResult DetectMethodMismatchResult
	UserAgentResult      DetectUserAgentAnomaliesResult
	ClientIPResult       DetectClientIPAnomaliesResult
	ASNResult            DetectASNAnomaliesResult
	NewEndpointResult    DetectNewEndpointsResult
	AuthResult           DetectAuthAnomaliesResult
	CleanupResult        CleanupStaleResult
}

// HttpMonitorAnomalyWorkflow orchestrates the sequential anomaly detection pipeline.
func HttpMonitorAnomalyWorkflow(ctx workflow.Context) (*HttpMonitorAnomalyResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting HttpMonitorAnomalyWorkflow")

	activityOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 15 * time.Minute,
		HeartbeatTimeout:    2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3,
		},
	}
	activityCtx := workflow.WithActivityOptions(ctx, activityOpts)

	result := &HttpMonitorAnomalyResult{}
	var detectionErrors []error

	// 1. Detect rate anomalies.
	if err := workflow.ExecuteActivity(activityCtx, DetectRateAnomaliesActivity).
		Get(ctx, &result.RateResult); err != nil {
		logger.Error("DetectRateAnomalies failed", "error", err)
		detectionErrors = append(detectionErrors, fmt.Errorf("rate anomalies: %w", err))
	} else {
		logger.Info("DetectRateAnomalies done",
			"spikes", result.RateResult.Spikes,
			"drops", result.RateResult.Drops,
			"responseSizeSpikes", result.RateResult.ResponseSizeSpike,
			"offHoursSpikes", result.RateResult.OffHoursSpike,
			"bulkDataExtraction", result.RateResult.BulkDataExtraction)
	}

	// 2. Detect error bursts.
	if err := workflow.ExecuteActivity(activityCtx, DetectErrorBurstsActivity).
		Get(ctx, &result.ErrorResult); err != nil {
		logger.Error("DetectErrorBursts failed", "error", err)
		detectionErrors = append(detectionErrors, fmt.Errorf("error bursts: %w", err))
	} else {
		logger.Info("DetectErrorBursts done",
			"errorBursts", result.ErrorResult.ErrorBursts,
			"5xxBursts", result.ErrorResult.FiveXXBursts)
	}

	// 3. Detect suspicious patterns.
	if err := workflow.ExecuteActivity(activityCtx, DetectSuspiciousPatternsActivity).
		Get(ctx, &result.SuspiciousResult); err != nil {
		logger.Error("DetectSuspiciousPatterns failed", "error", err)
		detectionErrors = append(detectionErrors, fmt.Errorf("suspicious patterns: %w", err))
	} else {
		logger.Info("DetectSuspiciousPatterns done",
			"scanners", result.SuspiciousResult.ScannerDetected,
			"ipFloods", result.SuspiciousResult.SingleIPFlood,
			"endpointEnum", result.SuspiciousResult.EndpointEnumeration,
			"lfi", result.SuspiciousResult.PathTraversal,
			"sqli", result.SuspiciousResult.SQLInjection,
			"rce", result.SuspiciousResult.CommandInjection,
			"xss", result.SuspiciousResult.XSSProbe,
			"ssrf", result.SuspiciousResult.SSRFProbe,
			"pagScraping", result.SuspiciousResult.PaginationScraping)
	}

	// 4. Detect method mismatches.
	if err := workflow.ExecuteActivity(activityCtx, DetectMethodMismatchActivity).
		Get(ctx, &result.MethodMismatchResult); err != nil {
		logger.Error("DetectMethodMismatch failed", "error", err)
		detectionErrors = append(detectionErrors, fmt.Errorf("method mismatch: %w", err))
	} else {
		logger.Info("DetectMethodMismatch done",
			"detected", result.MethodMismatchResult.Detected)
	}

	// 5. Detect user agent anomalies.
	if err := workflow.ExecuteActivity(activityCtx, DetectUserAgentAnomaliesActivity).
		Get(ctx, &result.UserAgentResult); err != nil {
		logger.Error("DetectUserAgentAnomalies failed", "error", err)
		detectionErrors = append(detectionErrors, fmt.Errorf("user agent anomalies: %w", err))
	} else {
		logger.Info("DetectUserAgentAnomalies done",
			"newUA", result.UserAgentResult.NewUA,
			"shareShift", result.UserAgentResult.ShareShift,
			"automated", result.UserAgentResult.AutomatedClient,
			"uaSpoofing", result.UserAgentResult.UASpoofing)
	}

	// 6. Detect client IP anomalies.
	if err := workflow.ExecuteActivity(activityCtx, DetectClientIPAnomaliesActivity).
		Get(ctx, &result.ClientIPResult); err != nil {
		logger.Error("DetectClientIPAnomalies failed", "error", err)
		detectionErrors = append(detectionErrors, fmt.Errorf("client IP anomalies: %w", err))
	} else {
		logger.Info("DetectClientIPAnomalies done",
			"newSourceIP", result.ClientIPResult.NewSourceIP,
			"geoShift", result.ClientIPResult.GeoShift,
			"externalOnInternal", result.ClientIPResult.ExternalOnInternal,
			"ipConcentration", result.ClientIPResult.IPConcentration,
			"sanctionedCountry", result.ClientIPResult.SanctionedCountry,
			"ipRotation", result.ClientIPResult.IPRotation)
	}

	// 7. Detect ASN anomalies.
	if err := workflow.ExecuteActivity(activityCtx, DetectASNAnomaliesActivity).
		Get(ctx, &result.ASNResult); err != nil {
		logger.Error("DetectASNAnomalies failed", "error", err)
		detectionErrors = append(detectionErrors, fmt.Errorf("ASN anomalies: %w", err))
	} else {
		logger.Info("DetectASNAnomalies done",
			"newASN", result.ASNResult.NewASN,
			"hostingProvider", result.ASNResult.HostingProvider,
			"asnConcentration", result.ASNResult.ASNConcentration)
	}

	// 8. Detect new endpoints.
	if err := workflow.ExecuteActivity(activityCtx, DetectNewEndpointsActivity).
		Get(ctx, &result.NewEndpointResult); err != nil {
		logger.Error("DetectNewEndpoints failed", "error", err)
		detectionErrors = append(detectionErrors, fmt.Errorf("new endpoints: %w", err))
	} else {
		logger.Info("DetectNewEndpoints done", "found", result.NewEndpointResult.NewEndpoints)
	}

	// 9. Detect auth anomalies.
	if err := workflow.ExecuteActivity(activityCtx, DetectAuthAnomaliesActivity).
		Get(ctx, &result.AuthResult); err != nil {
		logger.Error("DetectAuthAnomalies failed", "error", err)
		detectionErrors = append(detectionErrors, fmt.Errorf("auth anomalies: %w", err))
	} else {
		logger.Info("DetectAuthAnomalies done",
			"authFailure", result.AuthResult.AuthFailureBurst,
			"credStuffing", result.AuthResult.CredentialStuffing,
			"otpBrute", result.AuthResult.OTPBruteForce,
			"privEsc", result.AuthResult.PrivilegeEscalation,
			"resetFlood", result.AuthResult.PasswordResetFlood,
			"regAbuse", result.AuthResult.RegistrationAbuse,
			"rateLimit", result.AuthResult.RateLimitTriggered,
			"authSuccess", result.AuthResult.AuthSuccessAfterBurst)
	}

	// 10. Cleanup stale data (always runs regardless of detection errors).
	if err := workflow.ExecuteActivity(activityCtx, CleanupStaleActivity).
		Get(ctx, &result.CleanupResult); err != nil {
		logger.Error("CleanupStale failed", "error", err)
		detectionErrors = append(detectionErrors, fmt.Errorf("cleanup stale: %w", err))
	} else {
		logger.Info("CleanupStale done",
			"anomaliesDeleted", result.CleanupResult.AnomaliesDeleted,
			"silverTrafficDeleted", result.CleanupResult.SilverTrafficDeleted,
			"silverUADeleted", result.CleanupResult.SilverUADeleted,
			"silverIPDeleted", result.CleanupResult.SilverIPDeleted)
	}

	logger.Info("HttpMonitorAnomalyWorkflow complete")

	if err := errors.Join(detectionErrors...); err != nil {
		return result, err
	}
	return result, nil
}
