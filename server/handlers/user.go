package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/evanebb/regauth/auth/local"
	"github.com/evanebb/regauth/server/middleware"
	"github.com/evanebb/regauth/server/response"
	"github.com/evanebb/regauth/user"
	"github.com/evanebb/regauth/util"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

type userCtxKey struct{}

func RequireRole(l *slog.Logger, role user.Role) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			u, ok := middleware.AuthenticatedUserFromContext(r.Context())
			if !ok {
				l.ErrorContext(r.Context(), "could not parse user from request context")
				response.WriteJSONError(w, http.StatusUnauthorized, "authentication failed")
				return
			}

			if u.Role != role {
				response.WriteJSONError(w, http.StatusForbidden, "insufficient permission")
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func UserParser(l *slog.Logger, s user.Store) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			username := chi.URLParam(r, "username")
			if username == "" {
				response.WriteJSONError(w, http.StatusBadRequest, "no username given")
				return
			}

			u, err := s.GetByUsername(r.Context(), username)
			if err != nil {
				if errors.Is(err, user.ErrNotFound) {
					response.WriteJSONError(w, http.StatusNotFound, "user not found")
					return
				}

				l.ErrorContext(r.Context(), "could not get user", slog.Any("error", err))
				response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			ctx := withUser(r.Context(), u)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// withUser sets the given user.User in the context.
// Use userFromContext to retrieve the user.
func withUser(ctx context.Context, u user.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, u)
}

// userFromContext parses the current user.User from the given request context.
// This requires the user to have been previously set in the context, for example by the UserParser middleware.
func userFromContext(ctx context.Context) (user.User, bool) {
	val, ok := ctx.Value(userCtxKey{}).(user.User)
	return val, ok
}

func CreateUser(l *slog.Logger, s user.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newUser user.User
		if err := json.NewDecoder(r.Body).Decode(&newUser); err != nil {
			response.WriteJSONError(w, http.StatusBadRequest, "invalid JSON body given")
			return
		}

		_, err := s.GetByUsername(r.Context(), string(newUser.Username))
		if err == nil {
			response.WriteJSONError(w, http.StatusBadRequest, "user already exists")
			return
		}

		if !errors.Is(err, user.ErrNotFound) {
			l.ErrorContext(r.Context(), "could not get user", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		newUser.ID, err = uuid.NewV7()
		if err != nil {
			l.Error("could not generate UUID", "error", err)
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if err := newUser.IsValid(); err != nil {
			response.WriteJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := s.Create(r.Context(), newUser); err != nil {
			l.Error("could not create user", "error", err)
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		response.WriteJSONResponse(w, http.StatusOK, newUser)
	}
}

func ListUsers(l *slog.Logger, s user.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := s.GetAll(r.Context())
		if err != nil {
			l.ErrorContext(r.Context(), "could not get users", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		response.WriteJSONResponse(w, http.StatusOK, util.NilSliceToEmpty(users))
	}
}

func GetUser(l *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := userFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse user from request context")
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		response.WriteJSONResponse(w, http.StatusOK, u)
	}
}

func DeleteUser(l *slog.Logger, s user.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := middleware.AuthenticatedUserFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse authenticated user from request context")
			response.WriteJSONError(w, http.StatusUnauthorized, "authentication failed")
			return
		}

		u, ok := userFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse user from request context")
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if currentUser.ID == u.ID {
			response.WriteJSONError(w, http.StatusBadRequest, "cannot delete current user")
			return
		}

		if err := s.DeleteByID(r.Context(), u.ID); err != nil {
			l.ErrorContext(r.Context(), "could not delete user", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

type userPasswordChangeRequest struct {
	Password string `json:"password"`
}

func ChangeUserPassword(l *slog.Logger, s local.UserCredentialsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req userPasswordChangeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			response.WriteJSONError(w, http.StatusBadRequest, "invalid JSON body given")
			return
		}

		u, ok := userFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse user from request context")
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		credentials := local.UserCredentials{UserID: u.ID}
		if err := credentials.SetPassword(req.Password); err != nil {
			if errors.Is(err, local.ErrWeakPassword) {
				response.WriteJSONError(w, http.StatusBadRequest, err.Error())
				return
			}

			l.ErrorContext(r.Context(), "could not set password", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if err := s.Save(r.Context(), credentials); err != nil {
			l.ErrorContext(r.Context(), "could not update user credentials", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
