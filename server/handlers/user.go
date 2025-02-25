package handlers

import (
	"context"
	"errors"
	"github.com/evanebb/regauth/auth/local"
	"github.com/evanebb/regauth/httputil"
	"github.com/evanebb/regauth/session"
	"github.com/evanebb/regauth/template"
	"github.com/evanebb/regauth/user"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
	"strings"
)

type userCtxKey struct{}

// UserParser is a middleware that will look up the requested user.User from the ID in the path and sets it in the
// request context.
func UserParser(l *slog.Logger, t template.Templater, userStore user.Store) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
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

			ctx := withUser(r.Context(), u)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// withUser sets the given user.User in the context.
// Use userFromContext to retrieve the personal access token.
func withUser(ctx context.Context, t user.User) context.Context {
	return context.WithValue(ctx, userCtxKey{}, t)
}

// userFromContext parses the current user.User from the given request context.
// This requires the user to have been previously set in the context, for example by UserParser.
func userFromContext(ctx context.Context) (user.User, bool) {
	val, ok := ctx.Value(userCtxKey{}).(user.User)
	return val, ok
}

func UserOverview(l *slog.Logger, t template.Templater, userStore user.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := userStore.GetAll(r.Context())
		if err != nil {
			l.Error("failed to get users", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
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
		u, ok := userFromContext(r.Context())
		if !ok {
			l.Error("no requested user set in request context")
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		t.RenderBase(w, r, u, "account/users/view.gohtml")
	}
}

func CreateUserPage(t template.Templater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		t.RenderBase(w, r, nil, "account/users/create.gohtml")
	}
}

func CreateUser(l *slog.Logger, t template.Templater, userStore user.Store, authUserStore local.AuthUserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		u := user.User{
			ID:        uuid.New(),
			Username:  user.Username(r.PostFormValue("username")),
			FirstName: r.PostFormValue("firstName"),
			LastName:  r.PostFormValue("lastName"),
			Role:      user.Role(r.PostFormValue("role")),
		}

		err := u.IsValid()
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
		currentUser, ok := httputil.LoggedInUserFromContext(r.Context())
		if !ok {
			l.Error("no user in request context")
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		u, ok := userFromContext(r.Context())
		if !ok {
			l.Error("no requested user set in request context")
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		var err error
		if currentUser.ID == u.ID {
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

		err = userStore.DeleteByID(r.Context(), u.ID)
		if err != nil {
			l.Error("failed to delete user", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		err = authUserStore.DeleteByID(r.Context(), u.ID)
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
		u, ok := userFromContext(r.Context())
		if !ok {
			l.Error("no requested user set in request context")
			w.WriteHeader(http.StatusInternalServerError)
			t.RenderBase(w, r, nil, "errors/500.gohtml")
			return
		}

		authUser, err := authUserStore.GetByID(r.Context(), u.ID)
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

		http.Redirect(w, r, "/ui/account/users/"+u.ID.String(), http.StatusFound)
	}
}
