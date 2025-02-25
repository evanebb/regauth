package handlers

import (
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
)

func UserSessionParser(sessionStore sessions.Store, userStore user.Store) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			session, _ := sessionStore.Get(r, "session")
			val, ok := session.Values["userId"]
			if !ok {
				// If no user is found, do not attach it to the context
				next.ServeHTTP(w, r)
				return
			}

			raw, ok := val.(string)
			if !ok {
				// User ID isn't a string, no idea how we got here, just remove it from the session
				delete(session.Values, "userId")
				_ = session.Save(r, w)
				http.Redirect(w, r, "/ui/login", http.StatusFound)
				return
			}

			id, err := uuid.Parse(raw)
			if err != nil {
				// User ID is not a valid UUID, no idea how we got here, just remove it from the session
				delete(session.Values, "userId")
				_ = session.Save(r, w)
				http.Redirect(w, r, "/ui/login", http.StatusFound)
				return
			}

			u, err := userStore.GetByID(r.Context(), id)
			if err != nil {
				// User does not exist, remove the ID from the session so the user has to re-authenticate
				delete(session.Values, "userId")
				_ = session.Save(r, w)
				http.Redirect(w, r, "/ui/login", http.StatusFound)
				return
			}

			ctx := httputil.WithUser(r.Context(), u)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func UserAuth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if _, ok := httputil.UserFromContext(r.Context()); !ok {
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
			u, ok := httputil.UserFromContext(r.Context())
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
		err := r.ParseForm()
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}

		username := r.PostFormValue("username")
		password := r.PostFormValue("password")

		authUser, err := authUserStore.GetByUsername(r.Context(), username)
		if err != nil {
			if errors.Is(err, local.ErrUserNotFound) {
				l.Info("authentication failed: user does not exist", "username", username)
			} else {
				l.Error("authentication failed: failed to get auth user", "error", err, "username", username)
			}
			http.Error(w, "authentication failed", http.StatusUnauthorized)
			return
		}

		// This shouldn't ever happen, but handle it anyway
		u, err := userStore.GetByUsername(r.Context(), username)
		if err != nil {
			l.Error("authentication failed: failed to get user", "error", err, "username", username)
			http.Error(w, "authentication failed", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword(authUser.PasswordHash, []byte(password))
		if err != nil {
			l.Info("authentication failed: password does not match", "username", username)
			http.Error(w, "authentication failed", http.StatusUnauthorized)
			return
		}

		s, _ := sessionStore.Get(r, "session")
		s.Values["userId"] = u.ID.String()
		if u.Username == "admin" {
			s.AddFlash(session.NewFlash(session.FlashTypeWarning, "You are using the initial admin account. You should create a different admin account and delete this one."))
		}
		err = s.Save(r, w)
		if err != nil {
			l.Error("authentication failed: failed to save session", "error", err, "username", username)
			http.Error(w, "authentication failed", http.StatusUnauthorized)
			return
		}

		l.Info("successfully logged in user", "username", username)
		http.Redirect(w, r, "/ui", http.StatusFound)
	}
}

func Logout(l *slog.Logger, s sessions.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := s.Get(r, "session")
		delete(session.Values, "userId")
		err := session.Save(r, w)
		if err != nil {
			l.Error("failed to log out", "error", err)
			http.Error(w, "logout failed", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/ui", http.StatusFound)
	}
}

func LoginPage(t template.Templater) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, ok := httputil.UserFromContext(r.Context())
		if ok {
			http.Redirect(w, r, "/ui", http.StatusFound)
			return
		}

		w.WriteHeader(200)
		t.RenderBase(w, r, nil, "login.gohtml")
	}
}
