package storage

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rookgm/shortener/internal/db"
	"github.com/rookgm/shortener/internal/models"
)

// DBStorage presents database storage
type DBStorage struct {
	db *db.DataBase
}

// LoadFromFile does nothing
func (d *DBStorage) LoadFromFile() error { return nil }

// NewDBStorage creates a new storage on opened database
func NewDBStorage(db *db.DataBase) (*DBStorage, error) {
	return &DBStorage{db: db}, nil
}

// StoreURLCtx add url alias and original url to storage
func (d *DBStorage) StoreURLCtx(ctx context.Context, url models.ShrURL) error {
	stmt, err := d.db.DB.Prepare("INSERT INTO urls(url,alias) values($1,$2)")
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, url.URL, url.Alias)
	if err != nil {
		return err
	}
	defer stmt.Close()

	return nil
}

// GetURLCtx returns url alias and original url by alias
// if alias is not exist return an error
func (d *DBStorage) GetURLCtx(ctx context.Context, alias string) (models.ShrURL, error) {
	stmt, err := d.db.DB.Prepare("SELECT url FROM urls WHERE alias=$1")
	if err != nil {
		return models.ShrURL{}, err
	}

	var url string

	err = stmt.QueryRowContext(ctx, alias).Scan(&url)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.ShrURL{}, ErrURLNotFound
	case err != nil:
		return models.ShrURL{}, err
	}

	return models.ShrURL{Alias: alias, URL: url}, nil
}
