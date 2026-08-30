package util

import (
	"encoding/json"
	"net/http"
	"time"
)

// APIError is a typed, expected failure (bad input, not found, business
// rule violation). Handlers map these to HTTP status codes without ever
// leaking raw database errors to the client.
type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string { return e.Message }

func NewAPIError(status int, message string) *APIError {
	return &APIError{Status: status, Message: message}
}

func NotFound(what string) *APIError  { return NewAPIError(http.StatusNotFound, what+" not found") }
func BadRequest(msg string) *APIError { return NewAPIError(http.StatusBadRequest, msg) }
func Conflict(msg string) *APIError   { return NewAPIError(http.StatusConflict, msg) }
func UnprocessableEntity(msg string) *APIError {
	return NewAPIError(http.StatusUnprocessableEntity, msg)
}

type errorEnvelope struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// WriteJSON writes any payload as JSON with the given status code.
func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// WriteError converts an error into the standard CoreBank error envelope.
// Unknown errors (e.g. real database failures) are never exposed to the
// client — they are logged server-side and returned as a generic 500.
func WriteError(w http.ResponseWriter, err error) {
	if apiErr, ok := err.(*APIError); ok {
		WriteJSON(w, apiErr.Status, errorEnvelope{
			Success:   false,
			Message:   apiErr.Message,
			Timestamp: time.Now().Format(time.RFC3339),
		})
		return
	}

	WriteJSON(w, http.StatusInternalServerError, errorEnvelope{
		Success:   false,
		Message:   "Internal server error",
		Timestamp: time.Now().Format(time.RFC3339),
	})
}

type successEnvelope struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// WriteData wraps a successful payload in {"success": true, "data": ...}.
func WriteData(w http.ResponseWriter, status int, data interface{}) {
	WriteJSON(w, status, successEnvelope{Success: true, Data: data})
}

// DecodeJSON decodes a request body into dst, rejecting unknown fields so
// typos in the frontend are caught early instead of silently ignored.
func DecodeJSON(r *http.Request, dst interface{}) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return BadRequest("invalid request body: " + err.Error())
	}
	return nil
}
