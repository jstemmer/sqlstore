// Package pgdb provides a sqlstore.Database implementation for a PostgreSQL
// database accessed through the standard library database/sql package.
package pgdb

import (
	"context"
	"database/sql"
	"time"
)

// Database implements the sqlstore.Database interface for a PostgreSQL
// database.
type Database struct {
	db *sql.DB
}

// New creates a new pgdb.Database instance.
func New(db *sql.DB) *Database {
	return &Database{db}
}

// Load loads the session identified by id from the database.
func (d *Database) Load(ctx context.Context, id string) (updatedAt time.Time, data []byte, err error) {
	row := d.db.QueryRowContext(ctx, "SELECT data, updated_at FROM sessions WHERE id = $1 LIMIT 1", id)
	if err = row.Scan(&data, &updatedAt); err != nil {
		if err == sql.ErrNoRows {
			return time.Time{}, nil, nil
		}
		return time.Time{}, nil, err
	}
	return updatedAt, data, nil
}

// Insert saves a new session to the database.
func (d *Database) Insert(ctx context.Context, id string, data []byte) error {
	_, err := d.db.ExecContext(ctx, "INSERT INTO sessions(id, data, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())", id, data)
	return err
}

// Update updates an existing session in the database.
func (d *Database) Update(ctx context.Context, id string, data []byte) error {
	_, err := d.db.ExecContext(ctx, "UPDATE sessions SET data=$2, updated_at=NOW() WHERE id=$1", id, data)
	return err
}

// Delete deletes the session identified by id from the database.
func (d *Database) Delete(ctx context.Context, id string) error {
	_, err := d.db.ExecContext(ctx, "DELETE FROM sessions WHERE ID=$1", id)
	return err
}
