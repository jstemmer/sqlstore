// Package pgxdb provides a sqlstore.Database implementation for a PostgreSQL
// database accessed through the github.com/jackc/pgx package.
package pgxdb

import (
	"database/sql"
	"time"

	"github.com/jackc/pgx"
)

// Database implements the sqlstore.Database interface for a PostgreSQL
// database.
type Database struct {
	conn *pgx.Conn
}

// New creates a new pgxdb.Database instance with the given pgx connection
// conn.
func New(conn *pgx.Conn) *Database {
	return &Database{conn}
}

// Load loads the session identified by id from the database.
func (d *Database) Load(id string) (updatedAt time.Time, data []byte, err error) {
	row := d.conn.QueryRow("SELECT data, updated_at FROM sessions WHERE id = $1 LIMIT 1", id)
	if err = row.Scan(&data, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil, nil
		}
		return time.Time{}, nil, err
	}
	return updatedAt, data, nil
}

// Insert saves a new session to the database.
func (d *Database) Insert(id string, data []byte) error {
	_, err := d.conn.Exec("INSERT INTO sessions(id, data, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())", id, data)
	return err
}

// Update updates an existing session in the database.
func (d *Database) Update(id string, data []byte) error {
	_, err := d.conn.Exec("UPDATE sessions SET data=$2, updated_at=NOW() WHERE id=$1", id, data)
	return err
}

// Delete deletes the session identified by id from the database.
func (d *Database) Delete(id string) error {
	_, err := d.conn.Exec("DELETE FROM sessions WHERE ID=$1", id)
	return err
}
