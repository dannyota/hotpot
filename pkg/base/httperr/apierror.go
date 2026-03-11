package httperr

import "fmt"

// APIError represents an HTTP API error with a status code.
type APIError struct {
	Code int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error (status: %d)", e.Code)
}

// StatusCode returns the HTTP status code, satisfying the interface used by
// temporalerr.MaybeNonRetryable for non-retryable error detection.
func (e *APIError) StatusCode() int {
	return e.Code
}
