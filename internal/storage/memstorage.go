package storage

import (
	"context"
	"strings"
	"sync"

	"github.com/rookgm/shortener/internal/models"
)

// MemStorage is storage based on gomap
type MemStorage struct {
	mu sync.RWMutex
	m  map[string]string
	// user urls grouped by uid
	muser map[string][]models.ShrURL
}

// NewMemStorage creates a new storage in memory
func NewMemStorage() *MemStorage {
	return &MemStorage{
		m:     make(map[string]string),
		muser: make(map[string][]models.ShrURL),
	}
}

func (ms *MemStorage) LoadFromFile() error {
	// nothing
	return nil
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
	// put user url
	ms.muser[url.UserID] = append(ms.muser[url.UserID], url)
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
		// put user url
		ms.muser[url.UserID] = append(ms.muser[url.UserID], url)
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

// GetUserURLsCtx returns all user URLs by user ID
func (ms *MemStorage) GetUserURLsCtx(ctx context.Context, userID string) ([]models.ShrURL, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	urls, ok := ms.muser[userID]
	if !ok {
		return nil, ErrUserNotFound
	}

	return urls, nil
}

func (ms *MemStorage) DeleteUserURLsCtx(ctx context.Context, userID string, aliases []string) error {
	// TODO
	return nil
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
