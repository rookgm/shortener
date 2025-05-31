package storage

import (
	"errors"
	"github.com/rookgm/shortener/internal/models"
	"sync"
)

var ErrURLNotFound = errors.New("url not found")

// MemStorage is storage based on gomap
type MemStorage struct {
	mu sync.RWMutex
	m  map[string]string
}

// NewMemStorage creates a new storage in memory
func NewMemStorage() *MemStorage {
	return &MemStorage{
		m: make(map[string]string),
	}
}

// StoreURL is store ShrURL
func (ms *MemStorage) StoreURL(url models.ShrURL) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.m[url.Alias] = url.URL
	return nil
}

// GetURL is return ShrURL by alias
func (ms *MemStorage) GetURL(alias string) (models.ShrURL, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	url, ok := ms.m[alias]
	if !ok {
		return models.ShrURL{}, ErrURLNotFound
	}
	return models.ShrURL{Alias: alias, URL: url}, nil
}

func (ms *MemStorage) LoadFromFile() error {
	// nothing
	return nil
}
