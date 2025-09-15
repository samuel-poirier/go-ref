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
