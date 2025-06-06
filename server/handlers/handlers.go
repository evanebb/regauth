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
	"github.com/ogen-go/ogen/validate"
	"log/slog"
	"net/http"
)

func NotFound(w http.ResponseWriter, _ *http.Request) {
	response.WriteJSONError(w, http.StatusNotFound, "requested endpoint does not exist, please refer to the API documentation")
}

func MethodNotAllowed(w http.ResponseWriter, _ *http.Request, allowed string) {
	w.Header().Set("Allow", allowed)
	response.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed, please refer to the API documentation")
}

func ErrorHandler(l *slog.Logger) ogenerrors.ErrorHandler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, err error) {
		var (
			invalidContentTypeError *validate.InvalidContentTypeError
			validateError           *validate.Error
			decodeRequestError      *ogenerrors.DecodeRequestError
			decodeParamsError       *ogenerrors.DecodeParamsError
		)

		switch {
		case errors.As(err, &invalidContentTypeError):
			response.WriteJSONError(w, http.StatusUnsupportedMediaType, "unsupported content type: "+invalidContentTypeError.ContentType)
			return
		case errors.As(err, &validateError):
			response.WriteJSONError(w, http.StatusBadRequest, validateError.Error())
			return
		case errors.As(err, &decodeRequestError):
			l.DebugContext(ctx, "invalid request body", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusBadRequest, "invalid request body given")
			return
		case errors.As(err, &decodeParamsError):
			l.DebugContext(ctx, "invalid request params", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusBadRequest, "invalid request parameters given")
			return
		}

		// log the error and return a generic internal server error by default, to avoid potentially leaking sensitive info
		l.ErrorContext(ctx, "unhandled error occurred", slog.Any("error", err))
		response.WriteJSONInternalServerError(w)
	}
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
	tokenPrefix string,
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
			logger:      logger,
			tokenStore:  tokenStore,
			tokenPrefix: tokenPrefix,
		},
		UserHandler: UserHandler{
			logger:           logger,
			userStore:        userStore,
			credentialsStore: credentialsStore,
		},
	}
}

func (h Handler) NewError(ctx context.Context, err error) *oas.ErrorStatusCode {
	var (
		errorStatusCode *oas.ErrorStatusCode
		securityError   *ogenerrors.SecurityError
	)

	switch {
	case errors.As(err, &errorStatusCode):
		// if this is already a status code error, just pass it through
		// this should really only happen with errors returned from the SecurityHandler, since ogen will not check if
		// those are *oas.ErrorStatusCode instances
		return errorStatusCode
	case errors.Is(err, ogenerrors.ErrSecurityRequirementIsNotSatisfied), errors.As(err, &securityError):
		return newErrorResponse(http.StatusUnauthorized, "authentication failed")
	}

	// log the error and return a generic internal server error by default, to avoid potentially leaking sensitive info
	h.logger.ErrorContext(ctx, "unhandled error occurred", slog.Any("error", err))
	return newInternalServerErrorResponse()
}
