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
)

type URLStorage interface {
	StoreURLCtx(ctx context.Context, url models.ShrURL) error
	GetURLCtx(ctx context.Context, alias string) (models.ShrURL, error)
	GetAliasCtx(ctx context.Context, url string) (models.ShrURL, error)
	LoadFromFile() error
}
