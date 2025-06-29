package storage

import (
	"context"
	"github.com/rookgm/shortener/internal/models"
	"strings"
	"sync"
)

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

// isURLExist checks existing url
func (ms *MemStorage) isURLExist(url string) bool {
	// does the url exist?
	for _, v := range ms.m {
		if strings.Compare(v, url) == 0 {
			// url exist
			return true
		}
	}
	return false
}

// StoreURLCtx is store ShrURL
func (ms *MemStorage) StoreURLCtx(ctx context.Context, url models.ShrURL) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.isURLExist(url.URL) {
		return ErrURLExists
	}
	// put url
	ms.m[url.Alias] = url.URL
	return nil
}

// StoreBatchURLCtx stores batch urls
func (ms *MemStorage) StoreBatchURLCtx(ctx context.Context, urls []models.ShrURL) error {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for _, url := range urls {
		if ms.isURLExist(url.URL) {
			// url exists
			continue
		}
		// put url
		ms.m[url.Alias] = url.URL
	}
	return nil
}

// GetURLCtx is return ShrURL by alias
func (ms *MemStorage) GetURLCtx(ctx context.Context, alias string) (models.ShrURL, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	url, ok := ms.m[alias]
	if !ok {
		return models.ShrURL{}, ErrURLNotFound
	}
	return models.ShrURL{Alias: alias, URL: url}, nil
}

// GetAliasCtx returns stored alias by url
// if alias is not exist return an error
func (ms *MemStorage) GetAliasCtx(ctx context.Context, url string) (models.ShrURL, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	for k, v := range ms.m {
		if strings.Compare(v, url) == 0 {
			return models.ShrURL{Alias: k, URL: v}, nil
		}
	}
	return models.ShrURL{}, ErrAliasNotFound
}

func (ms *MemStorage) LoadFromFile() error {
	// nothing
	return nil
}
