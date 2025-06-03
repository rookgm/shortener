package db

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DataBase struct {
	DB *sql.DB
}

// OpenCtx opens a database using postgres driver
func OpenCtx(ctx context.Context, dsn string) (*DataBase, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	qu := `
		CREATE TABLE IF NOT EXISTS urls(
			id SERIAL PRIMARY KEY,
			alias TEXT NOT NULL,
			url TEXT NOT NULL UNIQUE);
		`

	// create table if not exist
	stmt, err := db.PrepareContext(ctx, qu)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	return &DataBase{DB: db}, nil
}

// Close closes the database
func (sdb *DataBase) Close() {
	sdb.DB.Close()
}

// PingCtx verifies a connection to the database
func (sdb *DataBase) PingCtx(ctx context.Context) error {
	return sdb.DB.PingContext(ctx)
}
