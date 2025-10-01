package storage

import (
	"context"
	"errors"

	"github.com/rookgm/shortener/internal/models"
)

// storage errors
var (
	// ErrURLNotFound is an error when URL is not found in the storage
	ErrURLNotFound = errors.New("url not found")
	// ErrURLExists is an error when URL is already exist in the storage
	ErrURLExists = errors.New("url exists")
	// ErrAliasNotFound is an error when shortened URL is not found in the storage
	ErrAliasNotFound = errors.New("alias not found")
	// ErrUserNotFound is an error when user's URL is not found in the storage
	ErrUserNotFound = errors.New("user not found")
)

// URLStorage is interface for interacting with storage-related data
type URLStorage interface {
	StoreURLCtx(ctx context.Context, url models.ShrURL) error
	StoreBatchURLCtx(ctx context.Context, urls []models.ShrURL) error
	GetURLCtx(ctx context.Context, alias string) (models.ShrURL, error)
	GetAliasCtx(ctx context.Context, url string) (models.ShrURL, error)
	GetUserURLsCtx(ctx context.Context, userID string) ([]models.ShrURL, error)
	DeleteUserURLsCtx(ctx context.Context, userID string, aliases []string) error
	LoadFromFile() error
	GetUserCountCtx(ctx context.Context) (int, error)
	GetURLCountCtx(ctx context.Context) (int, error)
}
