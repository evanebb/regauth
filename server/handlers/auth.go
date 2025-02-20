package handlers

import (
	"errors"
	"github.com/evanebb/regauth/auth/local"
	"github.com/evanebb/regauth/session"
	"github.com/evanebb/regauth/template"
	"github.com/evanebb/regauth/user"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
)

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
		_, err := getUserFromRequestContext(r.Context())
		if err == nil {
			http.Redirect(w, r, "/ui", http.StatusFound)
			return
		}

		w.WriteHeader(200)
		t.RenderBase(w, r, nil, "login.gohtml")
	}
}
