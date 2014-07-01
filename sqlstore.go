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
	db      *sql.DB
}

// New returns a new SQLStore. The keyPairs are used in the same way as the
// gorilla sessions CookieStore.
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
			if exists, err = s.load(session); err == nil && exists {
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

// load loads the session identified by its ID from the database. If the
// session has expired, it is destroyed. If no session could be found, exists
// will be false and no error will be returned.
func (s *SQLStore) load(session *sessions.Session) (exists bool, err error) {
	var data []byte
	var updatedAt time.Time
	row := s.db.QueryRow("SELECT data, updated_at FROM sessions WHERE id = $1 LIMIT 1", session.ID)
	if err := row.Scan(&data, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	if updatedAt.Add(time.Duration(s.Options.MaxAge) * time.Second).Before(time.Now().UTC()) {
		return false, s.destroy(session)
	}
	return true, gob.NewDecoder(bytes.NewBuffer(data)).Decode(&session.Values)
}

func (s *SQLStore) save(session *sessions.Session) error {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(session.Values); err != nil {
		return err
	}

	var query string
	if session.IsNew {
		query = "INSERT INTO sessions(id, data, created_at, updated_at) VALUES ($1, $2, NOW() AT TIME ZONE 'UTC', NOW() AT TIME ZONE 'UTC')"
	} else {
		query = "UPDATE sessions SET data=$2, updated_at=(NOW() AT TIME ZONE 'UTC') WHERE id=$1"
	}

	_, err := s.db.Exec(query, session.ID, buf.Bytes())
	return err
}

func (s *SQLStore) destroy(session *sessions.Session) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE ID=$1", session.ID)
	return err
}
