package middleware

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/auth/local"
	"github.com/evanebb/regauth/server/response"
	"github.com/evanebb/regauth/token"
	"github.com/evanebb/regauth/user"
	"log/slog"
	"net/http"
	"strings"
)

type ExcludedRoute struct {
	Path   string
	Method string
}

type ExcludedRoutes []ExcludedRoute

func (r ExcludedRoutes) IsExcluded(path, method string) bool {
	for _, route := range r {
		if route.Path == path && route.Method == method {
			return true
		}
	}

	return false
}

type authenticatedUserCtxKey struct{}

func TokenAuthentication(
	l *slog.Logger,
	tokenStore token.Store,
	userStore user.Store,
	excludedRoutes ExcludedRoutes,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			split := strings.Split(authHeader, " ")
			if len(split) != 2 || split[0] != "Bearer" {
				// allow bypassing token authentication for certain routes, it is assumed that authentication will be
				// handled separately
				// it is probably better to create some kind of authenticator interface + a stack of authenticators to
				// check, but eh, fine for now
				if excludedRoutes.IsExcluded(r.URL.Path, r.Method) {
					next.ServeHTTP(w, r)
					return
				}

				response.WriteJSONError(w, http.StatusUnauthorized, "authentication failed")
				return
			}

			t, err := tokenStore.GetByPlainTextToken(r.Context(), split[1])
			if err != nil {
				if errors.Is(err, token.ErrNotFound) {
					response.WriteJSONError(w, http.StatusUnauthorized, "authentication failed")
					return
				}

				l.ErrorContext(r.Context(), "could not get personal access token", slog.Any("error", err))
				response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			u, err := userStore.GetByID(r.Context(), t.UserID)
			if err != nil {
				l.ErrorContext(r.Context(), "could not get user for token", slog.Any("error", err),
					slog.String("userId", t.UserID.String()))
				response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			ctx := WithAuthenticatedUser(r.Context(), u)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

func UsernamePasswordAuthentication(l *slog.Logger, userStore user.Store, authUserStore local.AuthUserStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			// check if the user is already set in the request; if so, we do not have to do anything
			if _, ok := AuthenticatedUserFromContext(r.Context()); ok {
				next.ServeHTTP(w, r)
				return
			}

			username, password, ok := r.BasicAuth()
			if !ok {
				response.WriteJSONError(w, http.StatusUnauthorized, "authentication failed")
				return
			}

			authUser, err := authUserStore.GetByUsername(r.Context(), username)
			if err != nil {
				if errors.Is(err, local.ErrUserNotFound) {
					response.WriteJSONError(w, http.StatusUnauthorized, "authentication failed")
					return
				}

				l.ErrorContext(r.Context(), "could not get auth user", slog.Any("error", err))
				response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			if err := authUser.CheckPassword(password); err != nil {
				response.WriteJSONError(w, http.StatusUnauthorized, "authentication failed")
				return
			}

			u, err := userStore.GetByID(r.Context(), authUser.ID)
			if err != nil {
				// the user should always exist at this point, so this is always an error that should be logged
				l.ErrorContext(r.Context(), "could not get user", slog.Any("error", err))
				response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			ctx := WithAuthenticatedUser(r.Context(), u)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

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
