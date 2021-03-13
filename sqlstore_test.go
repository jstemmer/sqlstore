package sqlstore

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/sessions"
)

func TestImplementsGorillaSessionStore(t *testing.T) {
	var store sessions.Store
	store = New(nil, nil)
	_ = store
}

type testDatabase struct {
	t *testing.T

	calls  int
	action string
	id     string
	data   []byte
}

func (db *testDatabase) Load(ctx context.Context, id string) (updatedAt time.Time, data []byte, err error) {
	db.t.Logf("testDatabase.Load(%s)", id)
	db.calls++
	db.action = "select"
	db.id = id
	return time.Now(), db.data, nil
}

func (db *testDatabase) Insert(ctx context.Context, id string, data []byte) error {
	db.t.Logf("testDatabase.Insert(%s, [data:%d bytes])", id, len(data))
	db.calls++
	db.action = "insert"
	db.id = id
	db.data = data
	return nil
}

func (db *testDatabase) Update(ctx context.Context, id string, data []byte) error {
	db.t.Logf("testDatabase.Update(%s, %x)", id, data)
	db.calls++
	db.action = "update"
	db.id = id
	db.data = data
	return nil
}

func (db *testDatabase) Delete(ctx context.Context, id string) error {
	db.t.Logf("testDatabase.Delete(%s)", id)
	db.calls++
	db.action = "delete"
	db.id = id
	db.data = nil
	return nil
}

func (db *testDatabase) checkNotCalled() {
	db.t.Helper()
	db.checkCalled(0, "", "")
}

func (db *testDatabase) checkCalled(calls int, action, id string) {
	db.t.Helper()

	if db.calls != calls {
		db.t.Errorf("Incorrect database calls. Got %d, want %d", db.calls, calls)
	}

	if calls > 0 {
		if db.action != action {
			db.t.Errorf("Invalid database action. Got %s, want %s", db.action, action)
		}

		if db.id != id {
			db.t.Errorf("Invalid session ID. Got %s, want %s", db.id, id)
		}
	}
}

func setup(t *testing.T) (*SQLStore, *testDatabase, *http.Request) {
	t.Helper()

	generateSessionID = func() string { return "generated-session-id" }

	db := &testDatabase{t: t}
	store := New(db, []byte("testkey"))

	req, err := http.NewRequest("", "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	return store, db, req
}

func prepareSession(t *testing.T, sessionName string, values map[interface{}]interface{}) (*http.Cookie, []byte) {
	t.Helper()
	store, db, req := setup(t)

	session, err := store.New(req, sessionName)
	if err != nil {
		t.Fatal(err)
	}
	session.Values = values

	w := httptest.NewRecorder()
	if err = store.Save(req, w, session); err != nil {
		t.Fatal(err)
	}
	db.checkCalled(1, "insert", "generated-session-id")

	cookies := w.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("Incorrect number of expected cookies. Got %d, want %d", len(cookies), 1)
	}
	return cookies[0], db.data
}

func TestCreateNewSession(t *testing.T) {
	store, db, req := setup(t)

	session, err := store.Get(req, "session-name")
	if err != nil {
		t.Fatal(err)
	}
	if !session.IsNew {
		t.Errorf("session.IsNew is false, want true")
	}
	db.checkNotCalled()

	w := httptest.NewRecorder()
	if err = store.Save(req, w, session); err != nil {
		t.Fatal(err)
	}
	db.checkCalled(1, "insert", "generated-session-id")

	if len(w.HeaderMap["Set-Cookie"]) == 0 {
		t.Errorf("Expected a Set-Cookie header in the response.")
	}
}

func TestLoadExistingSession(t *testing.T) {
	store, db, req := setup(t)
	values := map[interface{}]interface{}{"key": "value"}

	// Prepare existing session
	var cookie *http.Cookie
	cookie, db.data = prepareSession(t, "existing-session-name", values)
	req.AddCookie(cookie)

	// Load session
	session, err := store.Get(req, "existing-session-name")
	if err != nil {
		t.Fatal(err)
	}
	db.checkCalled(1, "select", "generated-session-id")

	if session.IsNew {
		t.Errorf("session.IsNew is true, want false")
	}

	if diff := cmp.Diff(session.Values, values); diff != "" {
		t.Errorf("Invalid session.Values, diff (-got, +want):\n%s", diff)
	}
}

func TestUpdateSession(t *testing.T) {
	store, db, req := setup(t)

	session, err := store.Get(req, "session-name")
	session.ID = "test-session-1"
	session.IsNew = false

	w := httptest.NewRecorder()
	if err = store.Save(req, w, session); err != nil {
		t.Fatal(err)
	}
	db.checkCalled(1, "update", "test-session-1")

	if len(w.HeaderMap["Set-Cookie"]) == 0 {
		t.Errorf("Expected a Set-Cookie header in the response.")
	}
}

func TestDeleteSession(t *testing.T) {
	store, db, req := setup(t)

	session, err := store.Get(req, "session-name")
	if err != nil {
		t.Fatal(err)
	}

	session.ID = "test-session-1"
	session.IsNew = false
	session.Options.MaxAge = -1

	w := httptest.NewRecorder()
	if err = store.Save(req, w, session); err != nil {
		t.Fatal(err)
	}
	db.checkCalled(1, "delete", "test-session-1")
}
