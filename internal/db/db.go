package db

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DataBase struct {
	db *sql.DB
}

// Open opens a database using postgres driver
func Open(dataSourceName string) (*DataBase, error) {
	db, err := sql.Open("pgx", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &DataBase{db: db}, nil
}

// Close closes the database
func (sdb *DataBase) Close() {
	sdb.db.Close()
}

// PingCtx verifies a connection to the database
func (sdb *DataBase) PingCtx(ctx context.Context) error {
	return sdb.db.PingContext(ctx)
}
