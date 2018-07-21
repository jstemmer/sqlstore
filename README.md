# sqlstore

SQL session store for the [Gorilla web toolkit](http://www.gorillatoolkit.org).

Work in progress, API may not be stable. Currently only supports the PostgreSQL
dialect for either the stdlib database/sql package or the
[pgx](https://github.com/jackc/pgx) driver.

[![GoDoc](https://godoc.org/github.com/jstemmer/sqlstore?status.svg)](https://godoc.org/github.com/jstemmer/sqlstore)

[![Build Status](https://travis-ci.org/jstemmer/sqlstore.svg?branch=master)](https://travis-ci.org/jstemmer/sqlstore)

## Usage

See the [Gorilla toolkit sessions
documentation](http://www.gorillatoolkit.org/pkg/sessions) on how to use the
sessions package.

Get the package `go get github.com/jstemmer/sqlstore` and import
`github.com/jstemmer/sqlstore`. Depending on which database implementation you
want to use either import `github.com/jstemmer/sqlstore/pgdb` or
`github.com/jstemmer/sqlstore/pgxdb`.

Call `sqlstore.New()` to create a new store and provide it with a database
implementation and keyPairs. See the
[session.NewCookieStore](http://www.gorillatoolkit.org/pkg/sessions#NewCookieStore)
documentation on how to use the keyPairs parameter.

For example:
```go
store := sqlstore.New(pgdb.New(db), []byte("something-very-secret"))
```

Sessions are stored in a `sessions` table, which is assumed to exist. It should
have the following schema:

```sql
CREATE TABLE sessions (
	id varchar(100) PRIMARY KEY,
	data bytea NOT NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL
);
```

## License

MIT, see LICENSE.
