package handlers

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/auth/local"
	"github.com/evanebb/regauth/oas"
	"github.com/evanebb/regauth/repository"
	"github.com/evanebb/regauth/server/response"
	"github.com/evanebb/regauth/token"
	"github.com/evanebb/regauth/user"
	"github.com/ogen-go/ogen/ogenerrors"
	"log/slog"
	"net/http"
)

func NotFound(w http.ResponseWriter, r *http.Request) {
	response.WriteJSONError(w, http.StatusNotFound, "requested endpoint does not exist, please refer to the API documentation")
}

type Handler struct {
	logger *slog.Logger
	RepositoryHandler
	TeamHandler
	TokenHandler
	UserHandler
}

func NewHandler(
	logger *slog.Logger,
	repoStore repository.Store,
	userStore user.Store,
	teamStore user.TeamStore,
	tokenStore token.Store,
	credentialsStore local.UserCredentialsStore,
) Handler {
	return Handler{
		logger: logger,
		RepositoryHandler: RepositoryHandler{
			logger:    logger,
			repoStore: repoStore,
			teamStore: teamStore,
		},
		TeamHandler: TeamHandler{
			logger:    logger,
			teamStore: teamStore,
			userStore: userStore,
		},
		TokenHandler: TokenHandler{
			logger:     logger,
			tokenStore: tokenStore,
		},
		UserHandler: UserHandler{
			logger:           logger,
			userStore:        userStore,
			credentialsStore: credentialsStore,
		},
	}
}

func (h Handler) NewError(ctx context.Context, err error) *oas.ErrorStatusCode {
	switch {
	case errors.Is(err, ogenerrors.ErrSecurityRequirementIsNotSatisfied):
		// no credentials given
		return newErrorResponse(http.StatusUnauthorized, "authentication failed")
	}

	// log the error and return a generic internal server error by default, to avoid potentially leaking sensitive info
	h.logger.ErrorContext(ctx, "unhandled error occurred: "+err.Error(), slog.Any("error", err))
	return newErrorResponse(http.StatusInternalServerError, "internal server error")
}
