// Package pgxdb provides a sqlstore.Database implementation for a PostgreSQL
// database accessed through the github.com/jackc/pgx package.
package pgxdb

import (
	"context"
	"database/sql"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

// Database implements the sqlstore.Database interface for a PostgreSQL
// database.
type Database struct {
	pool *pgxpool.Pool
}

// New creates a new pgxdb.Database instance with the given pgx connection
// conn.
func New(pool *pgxpool.Pool) *Database {
	return &Database{pool}
}

// Load loads the session identified by id from the database.
func (d *Database) Load(ctx context.Context, id string) (updatedAt time.Time, data []byte, err error) {
	row := d.pool.QueryRow(ctx, "SELECT data, updated_at FROM sessions WHERE id = $1 LIMIT 1", id)
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
	_, err := d.pool.Exec(ctx, "INSERT INTO sessions(id, data, created_at, updated_at) VALUES ($1, $2, NOW(), NOW())", id, data)
	return err
}

// Update updates an existing session in the database.
func (d *Database) Update(ctx context.Context, id string, data []byte) error {
	_, err := d.pool.Exec(ctx, "UPDATE sessions SET data=$2, updated_at=NOW() WHERE id=$1", id, data)
	return err
}

// Delete deletes the session identified by id from the database.
func (d *Database) Delete(ctx context.Context, id string) error {
	_, err := d.pool.Exec(ctx, "DELETE FROM sessions WHERE ID=$1", id)
	return err
}
