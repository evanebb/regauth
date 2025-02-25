package handlers

import (
	"context"
	"errors"
	"fmt"
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
)

const userIdSessionKey = "userId"

func UserSessionParser(sessionStore sessions.Store, userStore user.Store) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			s, _ := sessionStore.Get(r, "session")
			val, ok := s.Values[userIdSessionKey]
			if !ok {
				// If no user is found, do not attach it to the context and continue serving the request
				next.ServeHTTP(w, r)
				return
			}

			u, err := getUserByRawID(r.Context(), val, userStore)
			if err != nil {
				// If we can't get a user using the supplied ID for some reason, remove it from the session so they have to re-authenticate
				delete(s.Values, userIdSessionKey)
				_ = s.Save(r, w)
				http.Redirect(w, r, "/ui/login", http.StatusFound)
				return
			}

			ctx := httputil.WithLoggedInUser(r.Context(), u)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func getUserByRawID(ctx context.Context, v interface{}, userStore user.Store) (user.User, error) {
	raw, ok := v.(string)
	if !ok {
		return user.User{}, errors.New("raw ID is not a string")
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		return user.User{}, fmt.Errorf("raw ID is not a valid UUID: %w", err)
	}

	u, err := userStore.GetByID(ctx, id)
	if err != nil {
		return user.User{}, fmt.Errorf("could not get user: %w", err)
	}

	return u, nil
}

func UserAuth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if _, ok := httputil.LoggedInUserFromContext(r.Context()); !ok {
			http.Redirect(w, r, "/ui/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func RequireRole(role user.Role) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			u, ok := httputil.LoggedInUserFromContext(r.Context())
			if !ok || u.Role != role {
				http.Redirect(w, r, "/ui", http.StatusFound)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func Login(l *slog.Logger, authUserStore local.AuthUserStore, userStore user.Store, sessionStore sessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, _ := sessionStore.Get(r, "session")
		defer func() {
			if err := s.Save(r, w); err != nil {
				l.Error("failed to save session", "error", err)
			}
		}()

		if err := r.ParseForm(); err != nil {
			s.AddFlash(session.NewFlash(session.FlashTypeError, "Bad request, please try again"))
			http.Redirect(w, r, "/ui/login", http.StatusFound)
			return
		}

		username := r.PostFormValue("username")
		password := r.PostFormValue("password")

		authUser, err := authUserStore.GetByUsername(r.Context(), username)
		if err != nil {
			var f session.Flash
			if errors.Is(err, local.ErrUserNotFound) {
				l.Info("authentication failed: user does not exist", "username", username)
				f = session.NewFlash(session.FlashTypeWarning, "Invalid username or password.")
			} else {
				l.Error("authentication failed: failed to get auth user", "error", err, "username", username)
				f = session.NewFlash(session.FlashTypeError, "Unknown error occurred.")
			}
			s.AddFlash(f)
			http.Redirect(w, r, "/ui/login", http.StatusFound)
			return
		}

		// This shouldn't ever happen, but handle it anyway
		u, err := userStore.GetByUsername(r.Context(), username)
		if err != nil {
			l.Error("authentication failed: failed to get user", "error", err, "username", username)
			s.AddFlash(session.NewFlash(session.FlashTypeError, "Unknown error occurred."))
			http.Redirect(w, r, "/ui/login", http.StatusFound)
			return
		}

		if err := bcrypt.CompareHashAndPassword(authUser.PasswordHash, []byte(password)); err != nil {
			l.Info("authentication failed: password does not match", "username", username)
			s.AddFlash(session.NewFlash(session.FlashTypeWarning, "Invalid username or password."))
			http.Redirect(w, r, "/ui/login", http.StatusFound)
			return
		}

		s.Values[userIdSessionKey] = u.ID.String()
		if u.Username == "admin" {
			s.AddFlash(session.NewFlash(session.FlashTypeWarning, "You are using the initial admin account. You should create a different admin account and delete this one."))
		}

		l.Info("successfully logged in user", "username", username)
		http.Redirect(w, r, "/ui", http.StatusFound)
	}
}

func Logout(l *slog.Logger, store sessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		s, _ := store.Get(r, "session")
		defer func() {
			if err := s.Save(r, w); err != nil {
				l.Error("failed to save session", "error", err)
			}
		}()

		delete(s.Values, userIdSessionKey)
		http.Redirect(w, r, "/ui", http.StatusFound)
	}
}

func LoginPage(t template.Templater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := httputil.LoggedInUserFromContext(r.Context())
		if ok {
			http.Redirect(w, r, "/ui", http.StatusFound)
			return
		}

		t.RenderBase(w, r, nil, "login.gohtml")
	}
}
