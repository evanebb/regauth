package handlers

import (
	"github.com/evanebb/regauth/server/response"
	"net/http"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	response.WriteJSONError(w, http.StatusNotFound, "requested endpoint does not exist, please refer to the API documentation")
}
