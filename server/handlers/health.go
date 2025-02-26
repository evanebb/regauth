package handlers

import "net/http"

// Health is a very simple handler that will always return a 200 OK status code.
// It does not do any other health-checking, only whether the server itself is up or not.
func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}
