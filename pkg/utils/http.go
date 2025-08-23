package utils

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func WriteJSON(w http.ResponseWriter, payload any, code int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(payload)
}

func DecodeBody(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// ValidationErrorResponse contains field-specific validation messages
// swagger:model ValidationErrorResponse
type ValidationErrorResponse struct {
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields"`
}

func WriteValidationError(w http.ResponseWriter, err error) error {
	res := ValidationErrorResponse{
		Message: "invalid request",
		Fields:  make(map[string]string),
	}

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fe := range ve {
			field := fe.Field()
			res.Fields[field] = fe.Tag()
		}
	}

	return WriteJSON(w, res, http.StatusBadRequest)
}

// ErrorResponse describes a standard error response
// swagger:model ErrorResponse
type ErrorResponse struct {
	Message string `json:"message"`
}

func WriteError(w http.ResponseWriter, message string, code int) error {
	return WriteJSON(w, ErrorResponse{Message: message}, code)
}
