package response

import (
	"encoding/json"
	"net/http"
)

type errorEnvelope struct {
	Message string `json:"message"`
}

func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSONResponse(w, statusCode, errorEnvelope{
		Message: message,
	})
}

func WriteJSONResponse(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}

func WriteJSONInternalServerError(w http.ResponseWriter) {
	WriteJSONError(w, http.StatusInternalServerError, "internal server error")
}
