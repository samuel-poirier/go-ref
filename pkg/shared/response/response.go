package response

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
)

func WriteJsonOk[T any](w http.ResponseWriter, payload T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(payload)
}

func WriteError(w http.ResponseWriter, err *ErrorModel) error {
	w.Header().Set("Content-Type", err.ContentType("application/json"))
	w.WriteHeader(err.Status)
	return json.NewEncoder(w).Encode(err)
}

func WriteJsonBadRequest(w http.ResponseWriter, errorMessage string, errors ...error) error {
	problemDetails := Error400BadRequest(errorMessage, errors...)
	problemDetails.Type = "https://datatracker.ietf.org/doc/html/rfc9110#section-15.5.1"
	return WriteError(w, problemDetails)
}

func WriteJsonBadRequestFromValidationErrors(w http.ResponseWriter, validation validator.ValidationErrors) error {
	errors := make([]error, len(validation))
	for i := range validation {
		errors[i] = validation[i]
	}
	return WriteJsonBadRequest(w, "validation errors", errors...)
}

func WriteInternalServerError(w http.ResponseWriter, errorMessage string) error {
	problemDetails := Error500InternalServerError(errorMessage)
	problemDetails.Type = "https://datatracker.ietf.org/doc/html/rfc9110#section-15.6.1"
	return WriteError(w, problemDetails)
}

func WriteNotFound(w http.ResponseWriter) error {
	problemDetails := Error404NotFound("object not found")
	problemDetails.Type = "https://datatracker.ietf.org/doc/html/rfc9110#section-15.5.5"
	return WriteError(w, problemDetails)
}
