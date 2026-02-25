package temporalerr

import (
	"errors"

	"danny.vn/greennode/sdkerror"
	"go.temporal.io/sdk/temporal"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MaybeNonRetryable checks whether err represents a permanent failure (e.g.
// permission denied, authentication failed) and, if so, wraps it with
// temporal.NewNonRetryableApplicationError so Temporal skips further retries.
// For all other errors the original error is returned unchanged.
func MaybeNonRetryable(err error) error {
	if err == nil {
		return nil
	}

	if reason := nonRetryableReason(err); reason != "" {
		return temporal.NewNonRetryableApplicationError(err.Error(), reason, err)
	}

	return err
}

// nonRetryableReason inspects err and returns a short error-type string if the
// error is known to be permanent, or "" if it should be retried normally.
func nonRetryableReason(err error) string {
	// GCP gRPC errors (PermissionDenied, Unauthenticated).
	if s, ok := status.FromError(err); ok {
		switch s.Code() {
		case codes.PermissionDenied:
			return "PERMISSION_DENIED"
		case codes.Unauthenticated:
			return "UNAUTHENTICATED"
		}
	}

	// GCP REST / googleapi errors (HTTP 401, 403).
	var gErr *googleapi.Error
	if errors.As(err, &gErr) {
		switch gErr.Code {
		case 401:
			return "UNAUTHENTICATED"
		case 403:
			return "PERMISSION_DENIED"
		}
	}

	// GreenNode SDK errors.
	var sdkErr *sdkerror.SdkError
	if errors.As(err, &sdkErr) {
		if sdkErr.IsErrorAny(sdkerror.EcPermissionDenied, sdkerror.EcAuthenticationFailed) {
			return "PERMISSION_DENIED"
		}
	}

	return ""
}
