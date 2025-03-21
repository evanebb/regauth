package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"net/http"
)

func getUUIDFromRequest(r *http.Request) (uuid.UUID, error) {
	u, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse UUID from request: %w", err)
	}

	return u, nil
}
