package response

import (
	"encoding/json"
	"net/http"
)

func WriteJsonOk[T any](w http.ResponseWriter, payload T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(payload)
}

func WriteJsonBadRequest[T any](w http.ResponseWriter, payload T) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	return json.NewEncoder(w).Encode(payload)
}

func WriteInternalServerError(w http.ResponseWriter, errorMessage string) (int, error) {
	w.Header().Set("Content-Type", "plain/text")
	w.WriteHeader(http.StatusInternalServerError)
	return w.Write([]byte(errorMessage))
}

func WriteNotFound(w http.ResponseWriter) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	return nil
}
