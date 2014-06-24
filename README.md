# sqlstore

SQL session store for the [Gorilla web toolkit](http://www.gorillatoolkit.org).

Work in progress, API may not be stable. Currently only supports the PostgreSQL
dialect.

[![Build Status](https://travis-ci.org/jstemmer/sqlstore.svg?branch=master)](https://travis-ci.org/jstemmer/sqlstore)

## Usage

See the [Gorilla toolkit sessions
documentation](http://www.gorillatoolkit.org/pkg/sessions) on how to use the
sessions package.

Get the package `go get github.com/jstemmer/sqlstore` and import
`github.com/jstemmer/sqlstore`.

Call `sqlstore.New()` to create a new store. You'll need a database handle from
the `database/sql` package. See the
[session.NewCookieStore](http://www.gorillatoolkit.org/pkg/sessions#NewCookieStore)
documentation on how to use the keyPairs parameter.

```go
store := sqlstore.New(db, []byte("something-very-secret"))
```

Sessions are stored in a `sessions` table, which is assumed to exist. It should
have the following schema:

```sql
CREATE TABLE sessions (
	id varchar(100) PRIMARY KEY,
	data text NOT NULL,
	created_at timestamp NOT NULL,
	updated_at timestamp NOT NULL
);
```

## Documentation

http://godoc.org/github.com/jstemmer/sqlstore

## License

MIT, see LICENSE.
