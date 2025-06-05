package postgres

import (
	"context"
	"encoding/base32"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
	"net/http"
)

type SessionStore struct {
	Codecs  []securecookie.Codec
	Options *sessions.Options
	db      *pgxpool.Pool
}

func NewSessionStore(db *pgxpool.Pool, keyPairs ...[]byte) *SessionStore {
	store := &SessionStore{
		Codecs: securecookie.CodecsFromPairs(keyPairs...),
		Options: &sessions.Options{
			HttpOnly: true,
			SameSite: http.SameSiteStrictMode,
			Path:     "/",
			MaxAge:   86400 * 30,
		},
		db: db,
	}

	store.MaxAge(store.Options.MaxAge)
	return store
}

func (s *SessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

func (s *SessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.Codecs...)
		if err == nil {
			err = s.load(session)
			if err == nil {
				session.IsNew = false
			}
		}
	}
	return session, err
}

func (s *SessionStore) Save(_ *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options.MaxAge <= 0 {
		err := s.erase(session)
		if err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return nil
	}

	if session.ID == "" {
		session.ID = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(securecookie.GenerateRandomKey(32))
	}

	err := s.save(session)
	if err != nil {
		return err
	}

	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.Codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

func (s *SessionStore) MaxAge(age int) {
	s.Options.MaxAge = age

	for _, codec := range s.Codecs {
		if sc, ok := codec.(*securecookie.SecureCookie); ok {
			sc.MaxAge(age)
		}
	}
}

func (s *SessionStore) save(session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, s.Codecs...)
	if err != nil {
		return err
	}

	if session.IsNew {
		query := "INSERT INTO sessions (session_id, data) VALUES ($1, $2)"
		_, err = s.db.Exec(context.Background(), query, session.ID, []byte(encoded))
		return err
	}

	query := "UPDATE sessions SET data = $1 WHERE session_id = $2"
	_, err = s.db.Exec(context.Background(), query, []byte(encoded), session.ID)
	return err
}

func (s *SessionStore) load(session *sessions.Session) error {
	var encoded []byte

	query := "SELECT data FROM sessions WHERE session_id = $1"
	err := s.db.QueryRow(context.Background(), query, session.ID).Scan(&encoded)
	if err != nil {
		return err
	}

	return securecookie.DecodeMulti(session.Name(), string(encoded), &session.Values, s.Codecs...)
}

func (s *SessionStore) erase(session *sessions.Session) error {
	query := "DELETE FROM sessions WHERE session_id = $1"
	_, err := s.db.Exec(context.Background(), query, session.ID)
	return err
}
