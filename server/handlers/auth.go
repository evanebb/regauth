package handlers

import (
	"encoding/json"
	"github.com/evanebb/regauth/auth/local"
	"github.com/evanebb/regauth/server/response"
	"github.com/evanebb/regauth/user"
	"github.com/go-faster/errors"
	"github.com/gorilla/sessions"
	"log/slog"
	"net/http"
)

const userIDSessionKey = "userID"

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func Login(
	l *slog.Logger,
	sessionStore sessions.Store,
	userStore user.Store,
	credentialsStore local.UserCredentialsStore,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var loginReq loginRequest
		if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
			response.WriteJSONError(w, http.StatusBadRequest, "invalid JSON body given")
			return
		}

		ctx := r.Context()
		u, err := userStore.GetByUsername(ctx, loginReq.Username)
		if err != nil {
			if errors.Is(err, user.ErrNotFound) {
				l.DebugContext(ctx, "user not found", slog.String("username", loginReq.Username))
				response.WriteJSONError(w, http.StatusUnauthorized, "authentication failed")
				return
			}

			l.ErrorContext(ctx, "could not get user", slog.Any("error", err))
			response.WriteJSONInternalServerError(w)
			return
		}

		credentials, err := credentialsStore.GetByUserID(ctx, u.ID)
		if err != nil {
			if errors.Is(err, local.ErrNoCredentials) {
				l.DebugContext(ctx, "no credentials set for user", slog.String("username", loginReq.Username))
				response.WriteJSONError(w, http.StatusUnauthorized, "authentication failed")
				return
			}

			l.ErrorContext(ctx, "could not get credentials", slog.Any("error", err))
			response.WriteJSONInternalServerError(w)
			return
		}

		if err := credentials.CheckPassword(loginReq.Password); err != nil {
			l.DebugContext(ctx, "password does not match", slog.String("username", loginReq.Username))
			response.WriteJSONError(w, http.StatusUnauthorized, "authentication failed")
			return
		}

		s, err := sessionStore.Get(r, "session")
		if err != nil {
			l.ErrorContext(ctx, "could not get session", slog.Any("error", err))
			response.WriteJSONInternalServerError(w)
			return
		}

		userIdString := u.ID.String()
		s.Values[userIDSessionKey] = userIdString
		if err := s.Save(r, w); err != nil {
			l.ErrorContext(ctx, "could not save session", slog.Any("error", err))
			response.WriteJSONInternalServerError(w)
			return
		}

		l.DebugContext(ctx, "login successful", slog.String("username", string(u.Username)), slog.String("userId", userIdString))
		w.WriteHeader(http.StatusNoContent)
	}
}

func Logout(
	l *slog.Logger,
	sessionStore sessions.Store,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		s, err := sessionStore.Get(r, "session")
		if err != nil {
			l.ErrorContext(ctx, "could not get session", slog.Any("error", err))
			response.WriteJSONInternalServerError(w)
			return
		}

		delete(s.Values, userIDSessionKey)
		if err := s.Save(r, w); err != nil {
			l.ErrorContext(ctx, "could not save session", slog.Any("error", err))
			response.WriteJSONInternalServerError(w)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
