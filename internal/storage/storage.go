package storage

import (
	"github.com/rookgm/shortener/internal/models"
)

type URLStorage interface {
	StoreURL(url models.ShrURL) error
	GetURL(alias string) (models.ShrURL, error)
	LoadFromFile() error
}
