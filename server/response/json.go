package response

import (
	"encoding/json"
	"net/http"
)

type responseEnvelope struct {
	Status  string `json:"status"`
	Data    any    `json:"data"`
	Message string `json:"message"`
}

func WriteJSONSuccess(w http.ResponseWriter, statusCode int, v any, message string) {
	WriteJSONResponse(w, statusCode, responseEnvelope{
		Status:  "success",
		Data:    v,
		Message: message,
	})
}

func WriteJSONError(w http.ResponseWriter, statusCode int, message string) {
	WriteJSONResponse(w, statusCode, responseEnvelope{
		Status:  "error",
		Message: message,
	})
}

func WriteJSONResponse(w http.ResponseWriter, statusCode int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(v)
}
