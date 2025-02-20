package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/evanebb/regauth/auth/local"
	"github.com/evanebb/regauth/pat"
	"github.com/evanebb/regauth/session"
	"github.com/evanebb/regauth/template"
	"github.com/evanebb/regauth/user"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

func ManageAccount(t template.Templater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.RenderBase(w, r, nil, "account/manage.gohtml")
	}
}

func TokenOverview(l *slog.Logger, t template.Templater, patStore pat.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := getUserFromRequestContext(r.Context())
		if err != nil {
			l.Error("failed to get user from request context", "error", err)
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
		u, err := getUserFromRequestContext(r.Context())
		if err != nil {
			l.Error("failed to get user from request context", "error", err)
			w.WriteHeader(500)
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
			exp, err = time.Parse("2006-01-02", customExp)
			if err != nil {
				l.Debug("invalid expiration date given", "error", err, "date", customExp)
				w.WriteHeader(http.StatusBadRequest)
				t.RenderBase(w, r, nil, "errors/400.gohtml")
				return
			}
			exp = exp.Add(24*time.Hour - time.Second)
		default:
			l.Debug("invalid expiration type given", "error", err, "expirationType", expirationType)
			w.WriteHeader(http.StatusBadRequest)
			t.RenderBase(w, r, nil, "errors/400.gohtml")
			return
		}

		randBytes := make([]byte, 42)
		_, _ = rand.Read(randBytes)
		plainTextToken := "registry_pat_" + base64.URLEncoding.EncodeToString(randBytes)
		hash, err := bcrypt.GenerateFromPassword([]byte(plainTextToken), bcrypt.DefaultCost)
		if err != nil {
			l.Error("failed to hash token", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		token := pat.PersonalAccessToken{
			ID:             uuid.New(),
			Hash:           hash,
			Description:    pat.Description(r.PostFormValue("description")),
			Permission:     pat.Permission(r.PostFormValue("permission")),
			ExpirationDate: exp,
			UserID:         u.ID,
		}

		err = token.IsValid()
		if err != nil {
			l.Debug("invalid personal access token given", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			t.RenderBase(w, r, nil, "errors/400.gohtml")
			return
		}

		err = patStore.Create(r.Context(), token)
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
		u, err := getUserFromRequestContext(r.Context())
		if err != nil {
			l.Error("failed to get user from request context", "error", err)
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

		usageLog, err := patStore.GetUsageLog(r.Context(), token.ID)
		if err != nil {
			l.Error("failed to get personal access token usage log", "error", err, "tokenId", id)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		paginated := PaginateRequest(r, usageLog, 5)

		data := struct {
			Token    pat.PersonalAccessToken
			UsageLog Pagination[[]pat.UsageLogEntry, pat.UsageLogEntry]
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
		u, err := getUserFromRequestContext(r.Context())
		if err != nil {
			l.Error("failed to get user from request context", "error", err)
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

		err = patStore.DeleteByID(r.Context(), token.ID)
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

func UserOverview(l *slog.Logger, t template.Templater, userStore user.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u, err := getUserFromRequestContext(r.Context())
		if err != nil {
			l.Error("failed to get user from request context", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		if u.Role != user.RoleAdmin {
			http.Redirect(w, r, "/ui", http.StatusFound)
			return
		}

		users, err := userStore.GetAll(r.Context())
		if err != nil {
			l.Error("failed to get users", "error", err)
			w.WriteHeader(500)
			t.Render(w, r, nil, "errors/500.gohtml")
			return
		}

		users = filterUsersByUsername(users, r.URL.Query().Get("q"))
		paginated := PaginateRequest(r, users, 10)

		if shouldRenderPartials(r) {
			t.Render(w, r, paginated, "partial", "account/users/overview.partial.gohtml")
		} else {
			t.RenderBase(w, r, paginated, "account/users/overview.gohtml", "account/users/overview.partial.gohtml")
		}
	}
}

func filterUsersByUsername(users []user.User, username string) []user.User {
	if username == "" {
		return users
	}

	var filtered []user.User
	for _, u := range users {
		if strings.Contains(u.Username.String(), username) {
			filtered = append(filtered, u)
		}
	}

	return filtered
}

func ViewUser(l *slog.Logger, t template.Templater, userStore user.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, err := getUserFromRequestContext(r.Context())
		if err != nil {
			l.Error("failed to get user from request context", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		if currentUser.Role != user.RoleAdmin {
			http.Redirect(w, r, "/ui", http.StatusFound)
			return
		}

		id, err := getUUIDFromRequest(r)
		if err != nil {
			l.Debug("could not get UUID from request", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			t.RenderBase(w, r, nil, "errors/400.gohtml")
			return
		}

		u, err := userStore.GetByID(r.Context(), id)
		if err != nil {
			if errors.Is(err, user.ErrNotFound) {
				l.Error("user not found", "error", err, "userId", id)
				w.WriteHeader(http.StatusNotFound)
				t.RenderBase(w, r, nil, "errors/404.gohtml")
				return
			}

			l.Error("failed to get user", "error", err, "userId", id)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		t.RenderBase(w, r, u, "account/users/view.gohtml")
	}
}

func CreateUserPage(l *slog.Logger, t template.Templater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, err := getUserFromRequestContext(r.Context())
		if err != nil {
			l.Error("failed to get user from request context", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		if currentUser.Role != user.RoleAdmin {
			http.Redirect(w, r, "/ui", http.StatusFound)
			return
		}

		t.RenderBase(w, r, nil, "account/users/create.gohtml")
	}
}

func CreateUser(l *slog.Logger, t template.Templater, userStore user.Store, authUserStore local.AuthUserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, err := getUserFromRequestContext(r.Context())
		if err != nil {
			l.Error("failed to get user from request context", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		if currentUser.Role != user.RoleAdmin {
			http.Redirect(w, r, "/ui", http.StatusFound)
			return
		}

		u := user.User{
			ID:        uuid.New(),
			Username:  user.Username(r.PostFormValue("username")),
			FirstName: r.PostFormValue("firstName"),
			LastName:  r.PostFormValue("lastName"),
			Role:      user.Role(r.PostFormValue("role")),
		}

		err = u.IsValid()
		if err != nil {
			l.Debug("invalid user given", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			t.RenderBase(w, r, nil, "errors/400.gohtml")
			return
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(r.PostFormValue("password")), bcrypt.DefaultCost)
		if err != nil {
			l.Error("failed to hash password", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		authUser := local.AuthUser{
			ID:           u.ID,
			Username:     u.Username.String(),
			PasswordHash: passwordHash,
		}

		// If one of the following calls returns an error, it can happen that the user has been created, but they have no credentials
		// A cross-store transaction (tied to the request context?) could solve this, but I don't care for now
		err = userStore.Create(r.Context(), u)
		if err != nil {
			l.Error("failed to create user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		// TODO: support multiple authentication back-ends (OAuth through GitHub/generic provider etc.)
		err = authUserStore.Create(r.Context(), authUser)
		if err != nil {
			l.Error("failed to create auth user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		http.Redirect(w, r, "/ui/account/users/"+u.ID.String(), http.StatusFound)
	}
}

func DeleteUser(l *slog.Logger, t template.Templater, userStore user.Store, authUserStore local.AuthUserStore, sessionStore sessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, err := getUserFromRequestContext(r.Context())
		if err != nil {
			l.Error("failed to get user from request context", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		if currentUser.Role != user.RoleAdmin {
			http.Redirect(w, r, "/ui", http.StatusFound)
			return
		}

		id, err := getUUIDFromRequest(r)
		if err != nil {
			l.Debug("could not get UUID from request", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			t.RenderBase(w, r, nil, "errors/400.gohtml")
			return
		}

		if currentUser.ID == id {
			s, _ := sessionStore.Get(r, "session")
			s.AddFlash(session.NewFlash(session.FlashTypeError, "Cannot delete currently logged-in user."))
			err = s.Save(r, w)
			if err != nil {
				l.Error("failed to save session", "error", err)
				http.Error(w, "authentication failed", http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/ui/account/users", http.StatusFound)
			return
		}

		err = userStore.DeleteByID(r.Context(), id)
		if err != nil {
			l.Error("failed to delete user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		err = authUserStore.DeleteByID(r.Context(), id)
		if err != nil {
			l.Error("failed to delete auth user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		s, _ := sessionStore.Get(r, "session")
		s.AddFlash(session.NewFlash(session.FlashTypeSuccess, "Successfully deleted user!"))
		err = s.Save(r, w)
		if err != nil {
			l.Error("failed to save session", "error", err)
			http.Error(w, "authentication failed", http.StatusUnauthorized)
			return
		}

		http.Redirect(w, r, "/ui/account/users", http.StatusFound)
	}
}

func ResetUserPassword(l *slog.Logger, t template.Templater, authUserStore local.AuthUserStore, sessionStore sessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, err := getUserFromRequestContext(r.Context())
		if err != nil {
			l.Error("failed to get user from request context", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		if currentUser.Role != user.RoleAdmin {
			http.Redirect(w, r, "/ui", http.StatusFound)
			return
		}

		id, err := getUUIDFromRequest(r)
		if err != nil {
			l.Debug("could not get UUID from request", "error", err)
			w.WriteHeader(http.StatusBadRequest)
			t.RenderBase(w, r, nil, "errors/400.gohtml")
			return
		}

		authUser, err := authUserStore.GetByID(r.Context(), id)
		if err != nil {
			l.Error("failed to get auth user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		newPassword := r.PostFormValue("newPassword")
		confirmPassword := r.PostFormValue("confirmPassword")

		if newPassword != confirmPassword {
			s, _ := sessionStore.Get(r, "session")
			s.AddFlash(session.NewFlash(session.FlashTypeError, "Passwords do not match."))
			err = s.Save(r, w)
			if err != nil {
				l.Error("failed to save session", "error", err)
				http.Error(w, "authentication failed", http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/ui/account/users", http.StatusFound)
			return
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
		if err != nil {
			l.Error("failed to hash password", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		authUser.PasswordHash = passwordHash
		err = authUserStore.Update(r.Context(), authUser)
		if err != nil {
			l.Error("failed to update auth user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		s, _ := sessionStore.Get(r, "session")
		s.AddFlash(session.NewFlash(session.FlashTypeSuccess, "Successfully reset password."))
		err = s.Save(r, w)
		if err != nil {
			l.Error("failed to save session", "error", err)
			http.Error(w, "authentication failed", http.StatusUnauthorized)
			return
		}

		http.Redirect(w, r, "/ui/account/users/"+id.String(), http.StatusFound)
	}
}
