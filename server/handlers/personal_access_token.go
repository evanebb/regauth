package handlers

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/evanebb/regauth/httputil"
	"github.com/evanebb/regauth/pat"
	"github.com/evanebb/regauth/session"
	"github.com/evanebb/regauth/template"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"log/slog"
	"net/http"
	"time"
)

type personalAccessTokenCtxKey struct{}

// PersonalAccessTokenParser is a middleware that will look up the requested pat.PersonalAccessToken from the ID in the
// path, checks if it belongs to the user and sets it in the request context.
func PersonalAccessTokenParser(l *slog.Logger, t template.Templater, patStore pat.Store) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			u, ok := httputil.LoggedInUserFromContext(r.Context())
			if !ok {
				l.Error("no user in request context")
				w.WriteHeader(http.StatusInternalServerError)
				t.RenderBase(w, r, nil, "errors/500.gohtml")
				return
			}

			id, err := getUUIDFromRequest(r)
			if err != nil {
				l.Debug("could not get UUID from request", "error", err)
				w.WriteHeader(http.StatusBadRequest)
				t.RenderBase(w, r, nil, "errors/400.gohtml")
				return
			}

			token, err := patStore.GetByID(r.Context(), id)
			if err != nil {
				if errors.Is(err, pat.ErrNotFound) {
					l.Error("personal access token not found", "error", err, "tokenId", id)
					w.WriteHeader(http.StatusNotFound)
					t.RenderBase(w, r, nil, "errors/404.gohtml")
					return
				}

				l.Error("failed to get personal access token", "error", err, "tokenId", id)
				w.WriteHeader(http.StatusInternalServerError)
				t.RenderBase(w, r, nil, "errors/500.gohtml")
				return
			}

			if token.UserID != u.ID {
				l.Debug("personal access token does not belong to user", "tokenId", token.ID, "userId", u.ID)
				w.WriteHeader(http.StatusNotFound)
				t.RenderBase(w, r, nil, "errors/400.gohtml")
				return
			}

			ctx := withPersonalAccessToken(r.Context(), token)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// withPersonalAccessToken sets the given pat.PersonalAccessToken in the context.
// Use personalAccessTokenFromContext to retrieve the personal access token.
func withPersonalAccessToken(ctx context.Context, t pat.PersonalAccessToken) context.Context {
	return context.WithValue(ctx, personalAccessTokenCtxKey{}, t)
}

// personalAccessTokenFromContext parses the current pat.PersonalAccessToken from the given request context.
// This requires the personal access token to have been previously set in the context, for example by PersonalAccessTokenParser.
func personalAccessTokenFromContext(ctx context.Context) (pat.PersonalAccessToken, bool) {
	val, ok := ctx.Value(personalAccessTokenCtxKey{}).(pat.PersonalAccessToken)
	return val, ok
}

func TokenOverview(l *slog.Logger, t template.Templater, patStore pat.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := httputil.LoggedInUserFromContext(r.Context())
		if !ok {
			l.Error("no user in request context")
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		tokens, err := patStore.GetAllForUser(r.Context(), u.ID)
		if err != nil {
			l.Error("failed to get personal access tokens for user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.Render(w, r, nil, "errors/500.gohtml")
		}

		t.RenderBase(w, r, tokens, "account/tokens/overview.gohtml")
	}
}

func CreateTokenPage(t template.Templater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.RenderBase(w, r, nil, "account/tokens/create.gohtml")
	}
}

func CreateToken(l *slog.Logger, t template.Templater, patStore pat.Store, registryHost string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, ok := httputil.LoggedInUserFromContext(r.Context())
		if !ok {
			l.Error("no user in request context")
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		var exp time.Time
		expirationType := r.PostFormValue("expirationType")
		switch expirationType {
		case "7d":
			exp = time.Now().AddDate(0, 0, 7)
		case "30d":
			exp = time.Now().AddDate(0, 0, 30)
		case "custom":
			customExp := r.PostFormValue("customExpirationDate")
			exp, err := time.Parse("2006-01-02", customExp)
			if err != nil {
				l.Debug("invalid expiration date given", "error", err, "date", customExp)
				w.WriteHeader(http.StatusBadRequest)
				t.RenderBase(w, r, nil, "errors/400.gohtml")
				return
			}
			exp = exp.Add(24*time.Hour - time.Second)
		default:
			l.Debug("invalid expiration type given", "expirationType", expirationType)
			w.WriteHeader(http.StatusBadRequest)
			t.RenderBase(w, r, nil, "errors/400.gohtml")
			return
		}

		randBytes := make([]byte, 42)
		_, _ = rand.Read(randBytes)
		plainTextToken := "registry_pat_" + base64.URLEncoding.EncodeToString(randBytes)

		token := pat.PersonalAccessToken{
			ID:             uuid.New(),
			Description:    pat.Description(r.PostFormValue("description")),
			Permission:     pat.Permission(r.PostFormValue("permission")),
			ExpirationDate: exp,
			UserID:         u.ID,
		}

		err := token.IsValid()
		if err != nil {
			l.Debug("invalid personal access token given", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			t.RenderBase(w, r, nil, "errors/400.gohtml")
			return
		}

		err = patStore.Create(r.Context(), token, plainTextToken)
		if err != nil {
			l.Error("failed to create token", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		data := struct {
			PlainTextToken string
			Token          pat.PersonalAccessToken
			RegistryHost   string
		}{
			PlainTextToken: plainTextToken,
			Token:          token,
			RegistryHost:   registryHost,
		}

		t.RenderBase(w, r, data, "account/tokens/generated.gohtml")
	}
}

func ViewToken(l *slog.Logger, t template.Templater, patStore pat.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := personalAccessTokenFromContext(r.Context())
		if !ok {
			l.Error("no personal access token set in request context")
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		usageLog, err := patStore.GetUsageLog(r.Context(), token.ID)
		if err != nil {
			l.Error("failed to get personal access token usage log", "error", err, "tokenId", token.ID)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		paginated := PaginateRequest(r, usageLog, 5)

		data := struct {
			Token    pat.PersonalAccessToken
			UsageLog Pagination[pat.UsageLogEntry]
		}{
			Token:    token,
			UsageLog: paginated,
		}

		if shouldRenderPartials(r) {
			t.Render(w, r, data, "partial", "account/tokens/view.partial.gohtml")
		} else {
			t.RenderBase(w, r, data, "account/tokens/view.gohtml", "account/tokens/view.partial.gohtml")
		}
	}
}

func DeleteToken(l *slog.Logger, t template.Templater, patStore pat.Store, sessionStore sessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := personalAccessTokenFromContext(r.Context())
		if !ok {
			l.Error("no personal access token set in request context")
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		err := patStore.DeleteByID(r.Context(), token.ID)
		if err != nil {
			l.Error("failed to delete personal access token", "error", err, "tokenId", token.ID)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		s, _ := sessionStore.Get(r, "session")
		s.AddFlash(session.NewFlash(session.FlashTypeSuccess, "Successfully deleted personal access token!"))
		err = s.Save(r, w)
		if err != nil {
			l.Error("failed to save session", "error", err)
			http.Error(w, "authentication failed", http.StatusUnauthorized)
			return
		}

		http.Redirect(w, r, "/ui/account/tokens", http.StatusFound)
	}
}
