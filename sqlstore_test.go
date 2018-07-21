package sqlstore

import (
	"testing"

	"github.com/gorilla/sessions"
)

func TestImplementsGorillaSessionStore(t *testing.T) {
	var store sessions.Store
	store = New(nil, nil)
	_ = store
}
