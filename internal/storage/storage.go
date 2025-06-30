package storage

import (
	"context"
	"errors"
	"github.com/rookgm/shortener/internal/models"
)

var (
	ErrURLNotFound   = errors.New("url not found")
	ErrURLExists     = errors.New("url exists")
	ErrAliasNotFound = errors.New("alias not found")
	ErrUserNotFound  = errors.New("user not found")
)

type URLStorage interface {
	StoreURLCtx(ctx context.Context, url models.ShrURL) error
	StoreBatchURLCtx(ctx context.Context, urls []models.ShrURL) error
	GetURLCtx(ctx context.Context, alias string) (models.ShrURL, error)
	GetAliasCtx(ctx context.Context, url string) (models.ShrURL, error)
	GetUserURLsCtx(ctx context.Context, userID string) ([]models.ShrURL, error)
	DeleteUserURLsCtx(ctx context.Context, userID string, aliases []string) error
	LoadFromFile() error
}
