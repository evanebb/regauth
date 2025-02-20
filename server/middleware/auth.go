package middleware

import (
	"context"
	"github.com/evanebb/regauth/user"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
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

			r = r.WithContext(context.WithValue(r.Context(), "user", u))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func UserAuth(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		u := r.Context().Value("user")
		if _, ok := u.(user.User); !ok {
			http.Redirect(w, r, "/ui/login", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
