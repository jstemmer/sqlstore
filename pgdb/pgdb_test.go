package pgdb

import (
	"testing"

	"github.com/jstemmer/sqlstore"
)

func TestImplementsSqlstoreDatabase(t *testing.T) {
	var _ sqlstore.Database = New(nil)
}
