package temporalerr

import (
	"errors"
	"fmt"
	"testing"

	"danny.vn/greennode/sdkerror"
	"go.temporal.io/sdk/temporal"
	"google.golang.org/api/googleapi"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMaybeNonRetryable_Nil(t *testing.T) {
	if got := MaybeNonRetryable(nil); got != nil {
		t.Fatalf("expected nil, got %v", got)
	}
}

func TestMaybeNonRetryable(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantType    string // expected error type, "" means retryable (passthrough)
		wantWrapped bool   // original error should be preserved as cause
	}{
		{
			name:        "gRPC PermissionDenied",
			err:         status.Error(codes.PermissionDenied, "caller lacks permission"),
			wantType:    "PERMISSION_DENIED",
			wantWrapped: true,
		},
		{
			name:        "gRPC Unauthenticated",
			err:         status.Error(codes.Unauthenticated, "bad credentials"),
			wantType:    "UNAUTHENTICATED",
			wantWrapped: true,
		},
		{
			name:        "wrapped gRPC PermissionDenied",
			err:         fmt.Errorf("ingest instances: %w", status.Error(codes.PermissionDenied, "no access")),
			wantType:    "PERMISSION_DENIED",
			wantWrapped: true,
		},
		{
			name:        "googleapi 403",
			err:         &googleapi.Error{Code: 403, Message: "forbidden"},
			wantType:    "PERMISSION_DENIED",
			wantWrapped: true,
		},
		{
			name:        "googleapi 401",
			err:         &googleapi.Error{Code: 401, Message: "unauthorized"},
			wantType:    "UNAUTHENTICATED",
			wantWrapped: true,
		},
		{
			name:        "wrapped googleapi 403",
			err:         fmt.Errorf("list buckets: %w", &googleapi.Error{Code: 403, Message: "forbidden"}),
			wantType:    "PERMISSION_DENIED",
			wantWrapped: true,
		},
		{
			name:        "greennode PermissionDenied",
			err:         sdkerror.NewPermissionDenied(),
			wantType:    "PERMISSION_DENIED",
			wantWrapped: true,
		},
		{
			name:        "wrapped greennode PermissionDenied",
			err:         fmt.Errorf("ingest servers: %w", sdkerror.NewPermissionDenied()),
			wantType:    "PERMISSION_DENIED",
			wantWrapped: true,
		},
		{
			name:     "gRPC NotFound is retryable",
			err:      status.Error(codes.NotFound, "resource not found"),
			wantType: "",
		},
		{
			name:     "gRPC Unavailable is retryable",
			err:      status.Error(codes.Unavailable, "service unavailable"),
			wantType: "",
		},
		{
			name:     "googleapi 500 is retryable",
			err:      &googleapi.Error{Code: 500, Message: "internal"},
			wantType: "",
		},
		{
			name:     "googleapi 429 is retryable",
			err:      &googleapi.Error{Code: 429, Message: "rate limited"},
			wantType: "",
		},
		{
			name:     "plain error is retryable",
			err:      errors.New("connection reset"),
			wantType: "",
		},
		{
			name:     "wrapped plain error is retryable",
			err:      fmt.Errorf("ingest: %w", errors.New("timeout")),
			wantType: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaybeNonRetryable(tt.err)
			if got == nil {
				t.Fatal("expected non-nil error")
			}

			var appErr *temporal.ApplicationError
			isNonRetryable := errors.As(got, &appErr)

			if tt.wantType == "" {
				// Should pass through unchanged.
				if isNonRetryable {
					t.Errorf("expected retryable error, got NonRetryableApplicationError type=%q", appErr.Type())
				}
				if got != tt.err {
					t.Errorf("expected same error instance returned")
				}
				return
			}

			// Should be non-retryable.
			if !isNonRetryable {
				t.Fatalf("expected NonRetryableApplicationError, got %T: %v", got, got)
			}
			if appErr.Type() != tt.wantType {
				t.Errorf("error type = %q, want %q", appErr.Type(), tt.wantType)
			}
			if !appErr.NonRetryable() {
				t.Error("expected NonRetryable() = true")
			}
			if tt.wantWrapped {
				if unwrapped := errors.Unwrap(appErr); unwrapped == nil {
					t.Error("expected wrapped cause, got nil")
				}
			}
		})
	}
}
