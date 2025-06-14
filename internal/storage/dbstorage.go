package storage

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
	stmt, err := d.db.DB.Prepare("INSERT INTO urls(userid,url,alias) values($1,$2,$3)")
	if err != nil {
		return err
	}
	_, err = stmt.ExecContext(ctx, url.UserID, url.URL, url.Alias)
	if err != nil {
		var pgErr *pgconn.PgError
		// does the url exist
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			// url exist
			return ErrURLExists
		} else {
			return err
		}
	}
	defer stmt.Close()

	return nil
}

// StoreBatchURLCtx stores batch urls
func (d *DBStorage) StoreBatchURLCtx(ctx context.Context, urls []models.ShrURL) error {
	tx, err := d.db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := d.db.DB.PrepareContext(ctx, "INSERT INTO urls(userid,url,alias) values($1,$2,$3)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, url := range urls {
		_, err = stmt.ExecContext(ctx, url.UserID, url.URL, url.Alias)
		if err != nil {
			var pgErr *pgconn.PgError
			// does the url exist
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				// url exist
				continue
			} else {
				return err
			}
		}
	}

	return tx.Commit()
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

// GetAliasCtx returns stored alias by url
// if alias is not exist return an error
func (d *DBStorage) GetAliasCtx(ctx context.Context, url string) (models.ShrURL, error) {
	stmt, err := d.db.DB.Prepare("SELECT alias FROM urls WHERE url=$1")
	if err != nil {
		return models.ShrURL{}, err
	}

	var alias string

	err = stmt.QueryRowContext(ctx, url).Scan(&alias)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return models.ShrURL{}, ErrURLNotFound
	case err != nil:
		return models.ShrURL{}, err
	}
	return models.ShrURL{Alias: alias, URL: url}, nil
}

// GetUserURLsCtx returns all user URLs by user ID
func (d *DBStorage) GetUserURLsCtx(ctx context.Context, userID string) ([]models.ShrURL, error) {
	rows, err := d.db.DB.Query("SELECT alias, url FROM urls WHERE userid=$1", userID)
	if err != nil {
		return nil, err
	}

	var userURLs []models.ShrURL

	for rows.Next() {
		var curURL models.ShrURL

		if err := rows.Scan(&curURL.Alias, &curURL.URL); err != nil {
			return nil, err
		}
		userURLs = append(userURLs, curURL)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return userURLs, nil
}
