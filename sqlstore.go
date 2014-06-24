package sqlstore

import (
	"database/sql"
	"encoding/base32"
	"net/http"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

type SQLStore struct {
	Options *sessions.Options
	codecs  []securecookie.Codec
	db      *sql.DB
}

func New(db *sql.DB, keyPairs ...[]byte) *SQLStore {
	return &SQLStore{
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		codecs: securecookie.CodecsFromPairs(keyPairs...),
		db:     db,
	}
}

func (s *SQLStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

func (s *SQLStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		if err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.codecs...); err == nil {
			if err = s.load(session); err == nil {
				session.IsNew = false
			}
		}
	}
	return session, err
}

func (s *SQLStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options.MaxAge < 0 {
		err := s.destroy(session)
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return err
	}

	if len(session.ID) == 0 {
		session.ID = base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
	}

	if err := s.save(session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

func (s *SQLStore) load(session *sessions.Session) error {
	var data string
	row := s.db.QueryRow("SELECT data FROM sessions WHERE id = $1 LIMIT 1", session.ID)
	if err := row.Scan(&data); err != nil {
		return err
	}
	return securecookie.DecodeMulti(session.Name(), data, &session.Values, s.codecs...)
}

func (s *SQLStore) save(session *sessions.Session) error {
	encoded, err := securecookie.EncodeMulti(session.Name(), session.Values, s.codecs...)
	if err != nil {
		return err
	}

	var query string
	if session.IsNew {
		query = "INSERT INTO sessions(id, data) VALUES ($1, $2)"
	} else {
		query = "UPDATE sessions SET data=$2 WHERE id=$1"
	}

	_, err = s.db.Exec(query, session.ID, encoded)
	return err
}

func (s *SQLStore) destroy(session *sessions.Session) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE ID=$1", session.ID)
	return err
}
