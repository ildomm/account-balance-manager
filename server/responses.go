package server

import (
	"encoding/json"
	"net/http"
)

// HealthResponse represents the response for the health check.
type HealthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

type UserResponse struct {
	UserID  int    `json:"userId"`
	Balance string `json:"balance"`
}

// ErrorResponse is the generic error API response container.
type ErrorResponse struct {
	Errors []string `json:"errors"`
}

// WriteInternalError writes a default internal error message as an HTTP response.
func WriteInternalError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(http.StatusText(http.StatusInternalServerError))) //nolint:all
}

// WriteErrorResponse takes an HTTP status code and a slice of errors
// and writes those as an HTTP error response in a structured format.
func WriteErrorResponse(w http.ResponseWriter, code int, errors []string) {
	w.WriteHeader(code)

	errorResponse := ErrorResponse{
		Errors: errors,
	}

	bytes, err := json.Marshal(errorResponse)
	if err != nil {
		WriteInternalError(w)
	}

	w.Write(bytes) //nolint:all
}

// WriteAPIResponse takes an HTTP status code and a generic data struct
// and writes those as an HTTP response in a structured format.
func WriteAPIResponse(w http.ResponseWriter, code int, data interface{}) {
	w.WriteHeader(code)

	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		WriteInternalError(w)
	}

	w.Write(bytes) //nolint:all
}
