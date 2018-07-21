package sqlstore

import (
	"bytes"
	"database/sql"
	"encoding/base32"
	"encoding/gob"
	"net/http"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
)

// SQLStore stores gorilla sessions in a database.
type SQLStore struct {
	Options *sessions.Options
	codecs  []securecookie.Codec
	db      Database
}

// Database describes a session store database interface.
type Database interface {
	Load(id string) (updatedAt time.Time, data []byte, err error)
	Insert(id string, data []byte) error
	Update(id string, data []byte) error
	Delete(id string) error
}

var generateSessionID = func() string {
	return base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32))
}

// New returns a new SQLStore. The keyPairs are used in the same way as the
// gorilla sessions CookieStore.
func New(db Database, keyPairs ...[]byte) *SQLStore {
	return &SQLStore{
		Options: &sessions.Options{
			Path:   "/",
			MaxAge: 86400 * 30,
		},
		codecs: securecookie.CodecsFromPairs(keyPairs...),
		db:     db,
	}
}

// Get returns a cached session.
func (s *SQLStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(s, name)
}

// New creates a new session for the given request r. If the request contains a
// valid session ID for an existing, non-expired session, then this session
// will be loaded from the database.
func (s *SQLStore) New(r *http.Request, name string) (*sessions.Session, error) {
	session := sessions.NewSession(s, name)
	opts := *s.Options
	session.Options = &opts
	session.IsNew = true
	var err error
	if c, errCookie := r.Cookie(name); errCookie == nil {
		if err = securecookie.DecodeMulti(name, c.Value, &session.ID, s.codecs...); err == nil {
			var exists bool
			if exists, err = s.loadFromDatabase(session); err == nil && exists {
				session.IsNew = false
			}
		}
	}
	return session, err
}

// Save stores the session in the database. If session.Options.MaxAge is < 0,
// the session is deleted from the database.
func (s *SQLStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options.MaxAge < 0 {
		err := s.db.Delete(session.ID)
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
		return err
	}

	if len(session.ID) == 0 {
		session.ID = generateSessionID()
	}

	if err := s.saveToDatabase(session); err != nil {
		return err
	}
	encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, s.codecs...)
	if err != nil {
		return err
	}
	http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
	return nil
}

// loadFromDatabase loads the session identified by its ID from the database.
// If the session has expired, it is destroyed. If no session could be found,
// exists will be false and no error will be returned.
func (s *SQLStore) loadFromDatabase(session *sessions.Session) (exists bool, err error) {
	updatedAt, data, err := s.db.Load(session.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if updatedAt.Add(time.Duration(s.Options.MaxAge) * time.Second).Before(time.Now().UTC()) {
		return false, s.db.Delete(session.ID)
	}
	return true, gob.NewDecoder(bytes.NewBuffer(data)).Decode(&session.Values)
}

// saveToDatabase stores session in the database.
func (s *SQLStore) saveToDatabase(session *sessions.Session) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(session.Values); err != nil {
		return err
	}
	if !session.IsNew {
		return s.db.Update(session.ID, buf.Bytes())
	}
	return s.db.Insert(session.ID, buf.Bytes())
}
