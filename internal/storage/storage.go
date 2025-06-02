package storage

import (
	"context"
	"github.com/rookgm/shortener/internal/models"
)

type URLStorage interface {
	StoreURLCtx(ctx context.Context, url models.ShrURL) error
	GetURLCtx(ctx context.Context, alias string) (models.ShrURL, error)
	LoadFromFile() error
}
