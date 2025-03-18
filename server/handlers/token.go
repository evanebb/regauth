package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/evanebb/regauth/server/middleware"
	"github.com/evanebb/regauth/server/response"
	"github.com/evanebb/regauth/token"
	"github.com/google/uuid"
	"log/slog"
	"net/http"
)

type personalAccessTokenCtxKey struct{}

func PersonalAccessTokenParser(l *slog.Logger, s token.Store) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			u, ok := middleware.AuthenticatedUserFromContext(r.Context())
			if !ok {
				l.ErrorContext(r.Context(), "could not parse user from request context")
				response.WriteJSONError(w, http.StatusUnauthorized, "authentication failed")
				return
			}

			id, err := getUUIDFromRequest(r)
			if err != nil {
				response.WriteJSONError(w, http.StatusBadRequest, "invalid ID given")
				return
			}

			pat, err := s.GetByID(r.Context(), id)
			if err != nil {
				if errors.Is(err, token.ErrNotFound) {
					response.WriteJSONError(w, http.StatusNotFound, "personal access token not found")
					return
				}

				l.ErrorContext(r.Context(), "could not get personal access token", slog.Any("error", err))
				response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			if pat.UserID != u.ID {
				response.WriteJSONError(w, http.StatusNotFound, "personal access token not found")
				return
			}

			ctx := withPersonalAccessToken(r.Context(), pat)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
		return http.HandlerFunc(fn)
	}
}

// withPersonalAccessToken sets the given token.PersonalAccessToken in the context.
// Use personalAccessTokenFromContext to retrieve the token.
func withPersonalAccessToken(ctx context.Context, t token.PersonalAccessToken) context.Context {
	return context.WithValue(ctx, personalAccessTokenCtxKey{}, t)
}

// personalAccessTokenFromContext parses the current token.PersonalAccessToken from the given request context.
// This requires the token to have been previously set in the context, for example by the PersonalAccessTokenParser middleware.
func personalAccessTokenFromContext(ctx context.Context) (token.PersonalAccessToken, bool) {
	val, ok := ctx.Value(personalAccessTokenCtxKey{}).(token.PersonalAccessToken)
	return val, ok
}

type tokenCreationResponse struct {
	token.PersonalAccessToken
	Token string `json:"token"`
}

func CreateToken(l *slog.Logger, s token.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := middleware.AuthenticatedUserFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse user from request context")
			response.WriteJSONError(w, http.StatusUnauthorized, "authenticated failed")
			return
		}

		var pat token.PersonalAccessToken
		if err := json.NewDecoder(r.Body).Decode(&pat); err != nil {
			response.WriteJSONError(w, http.StatusBadRequest, "invalid JSON body given")
			return
		}

		pat.ID = uuid.New()
		pat.UserID = u.ID

		if err := pat.IsValid(); err != nil {
			response.WriteJSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		randBytes := make([]byte, 42)
		_, _ = rand.Read(randBytes)
		plainTextToken := "registry_pat_" + base64.URLEncoding.EncodeToString(randBytes)

		if err := s.Create(r.Context(), pat, plainTextToken); err != nil {
			l.ErrorContext(r.Context(), "could not create personal access token", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		resp := tokenCreationResponse{PersonalAccessToken: pat, Token: plainTextToken}
		response.WriteJSONSuccess(w, http.StatusOK, resp, "successfully created token")
	}
}

func ListTokens(l *slog.Logger, s token.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := middleware.AuthenticatedUserFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse user from request context")
			response.WriteJSONError(w, http.StatusUnauthorized, "authentication failed")
			return
		}

		tokens, err := s.GetAllByUser(r.Context(), u.ID)
		if err != nil {
			l.ErrorContext(r.Context(), "could not get personal access tokens for user", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		response.WriteJSONSuccess(w, http.StatusOK, tokens, "successfully listed tokens")
	}
}

func GetToken(l *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pat, ok := personalAccessTokenFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse personal access token from request context")
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		response.WriteJSONSuccess(w, http.StatusOK, pat, "successfully returned personal access token")
	}
}

func DeleteToken(l *slog.Logger, s token.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		pat, ok := personalAccessTokenFromContext(r.Context())
		if !ok {
			l.ErrorContext(r.Context(), "could not parse personal access token from request context")
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if err := s.DeleteByID(r.Context(), pat.ID); err != nil {
			l.ErrorContext(r.Context(), "could not delete personal access token", slog.Any("error", err))
			response.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		response.WriteJSONSuccess(w, http.StatusOK, nil, "successfully deleted personal access token")
	}
}
