package sqlstore

import (
	"github.com/gorilla/sessions"
	"testing"
)

func TestImplementsGorillaSessionStore(t *testing.T) {
	var store sessions.Store = New(nil, nil)
	_ = store
}
