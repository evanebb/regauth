package handlers

import (
	"context"
	"errors"
	"fmt"
	"github.com/evanebb/regauth/user"
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

func getUserFromRequestContext(ctx context.Context) (user.User, error) {
	val := ctx.Value("user")
	if val == nil {
		return user.User{}, errors.New("no user set in request context")
	}

	u, ok := val.(user.User)
	if !ok {
		return user.User{}, errors.New("user set in request context is not valid")
	}

	return u, nil
}

func shouldRenderPartials(r *http.Request) bool {
	isHtmxRequest := r.Header.Get("HX-Request") != ""
	isBoosted := r.Header.Get("HX-Boosted") != ""
	return isHtmxRequest && !isBoosted
}
