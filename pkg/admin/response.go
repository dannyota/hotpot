package admin

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// FilterOption is a value with its count for filter dropdowns.
type FilterOption struct {
	Value string `json:"value"`
	Count int    `json:"count"`
}

// ListResponse is the standard envelope for paginated list endpoints.
type ListResponse struct {
	Data          any                        `json:"data"`
	Meta          PaginationMeta             `json:"meta"`
	FilterOptions map[string][]FilterOption  `json:"filter_options,omitempty"`
}

// PaginationMeta contains pagination info for list responses.
type PaginationMeta struct {
	Page       int `json:"page"`
	Size       int `json:"size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// DetailResponse is the standard envelope for detail endpoints.
type DetailResponse struct {
	Data any `json:"data"`
}

// ErrorResponse is the standard envelope for error responses.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error code and message.
type ErrorDetail struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// WriteError writes a JSON error response.
func WriteError(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, ErrorResponse{Error: ErrorDetail{Code: status, Message: msg}})
}

// WriteServerError logs the real error and writes a generic message to the client.
func WriteServerError(w http.ResponseWriter, msg string, err error) {
	slog.Error(msg, "error", err)
	WriteError(w, http.StatusInternalServerError, msg)
}
