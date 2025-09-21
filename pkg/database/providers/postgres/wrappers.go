// Package postgres implements wrapper types to match database interfaces
package postgres

import (
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// PostgresRows wraps pgx.Rows to implement interfaces.Rows
type PostgresRows struct {
	rows pgx.Rows
}

func (r *PostgresRows) Next() bool {
	return r.rows.Next()
}

func (r *PostgresRows) Scan(dest ...interface{}) error {
	return r.rows.Scan(dest...)
}

func (r *PostgresRows) Close() error {
	r.rows.Close()
	return nil
}

func (r *PostgresRows) Err() error {
	return r.rows.Err()
}

// PostgresRow wraps pgx.Row to implement interfaces.Row
type PostgresRow struct {
	row pgx.Row
	err error
}

func (r *PostgresRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	return r.row.Scan(dest...)
}

// PostgresResult wraps pgconn.CommandTag to implement interfaces.Result
type PostgresResult struct {
	result pgconn.CommandTag
}

func (r *PostgresResult) LastInsertId() (int64, error) {
	// PostgreSQL doesn't return last insert ID by default
	// This would need to be handled differently (e.g., using RETURNING clause)
	return 0, nil
}

func (r *PostgresResult) RowsAffected() (int64, error) {
	return r.result.RowsAffected(), nil
}
