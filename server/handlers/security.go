package handlers

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/auth/local"
	"github.com/evanebb/regauth/oas"
	"github.com/evanebb/regauth/token"
	"github.com/evanebb/regauth/user"
	"log/slog"
	"net/http"
)

type SecurityHandler struct {
	logger           *slog.Logger
	tokenStore       token.Store
	userStore        user.Store
	credentialsStore local.UserCredentialsStore
}

func NewSecurityHandler(
	logger *slog.Logger,
	tokenStore token.Store,
	userStore user.Store,
	credentialsStore local.UserCredentialsStore,
) SecurityHandler {
	return SecurityHandler{
		logger:           logger,
		tokenStore:       tokenStore,
		userStore:        userStore,
		credentialsStore: credentialsStore,
	}
}

func (s SecurityHandler) HandlePersonalAccessToken(ctx context.Context, operationName oas.OperationName, t oas.PersonalAccessToken) (context.Context, error) {
	tok, err := s.tokenStore.GetByPlainTextToken(ctx, t.GetToken())
	if err != nil {
		if errors.Is(err, token.ErrNotFound) {
			s.logger.DebugContext(ctx, "token does not exist")
			return ctx, newErrorResponse(http.StatusUnauthorized, "authentication failed")
		}

		s.logger.ErrorContext(ctx, "could not get personal access token", slog.Any("error", err))
		return ctx, newInternalServerErrorResponse()
	}

	u, err := s.userStore.GetByID(ctx, tok.UserID)
	if err != nil {
		return ctx, newInternalServerErrorResponse()
	}

	s.logger.DebugContext(ctx, "token authentication successful")
	return WithAuthenticatedUser(ctx, u), nil
}

func (s SecurityHandler) HandleUsernamePassword(ctx context.Context, operationName oas.OperationName, t oas.UsernamePassword) (context.Context, error) {
	u, err := s.userStore.GetByUsername(ctx, t.GetUsername())
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			s.logger.DebugContext(ctx, "user not found", slog.String("username", t.GetUsername()))
			return ctx, newErrorResponse(http.StatusUnauthorized, "authentication failed")
		}

		s.logger.ErrorContext(ctx, "could not get user", slog.Any("error", err))
		return ctx, newInternalServerErrorResponse()
	}

	credentials, err := s.credentialsStore.GetByUserID(ctx, u.ID)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			s.logger.DebugContext(ctx, "no credentials set for user", slog.String("username", t.GetUsername()))
			return ctx, newErrorResponse(http.StatusUnauthorized, "authentication failed")
		}

		s.logger.ErrorContext(ctx, "could not get credentials", slog.Any("error", err))
		return ctx, newInternalServerErrorResponse()
	}

	if err := credentials.CheckPassword(t.GetPassword()); err != nil {
		s.logger.DebugContext(ctx, "password does not match", slog.String("username", t.GetUsername()))
		return ctx, newErrorResponse(http.StatusUnauthorized, "authentication failed")
	}

	s.logger.DebugContext(ctx, "token authentication successful")
	return WithAuthenticatedUser(ctx, u), nil
}

type authenticatedUserCtxKey struct{}

// WithAuthenticatedUser sets the authenticated user.User in the context.
// Use AuthenticatedUserFromContext to retrieve the user.
func WithAuthenticatedUser(ctx context.Context, u user.User) context.Context {
	return context.WithValue(ctx, authenticatedUserCtxKey{}, u)
}

// AuthenticatedUserFromContext parses the authenticated user.User from the given request context.
// This requires the user to have been previously set in the context, for example by the TokenAuthentication middleware.
func AuthenticatedUserFromContext(ctx context.Context) (user.User, bool) {
	val, ok := ctx.Value(authenticatedUserCtxKey{}).(user.User)
	return val, ok
}
