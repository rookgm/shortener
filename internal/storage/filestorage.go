package storage

import (
	"context"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/rookgm/shortener/internal/models"
	"github.com/rookgm/shortener/internal/recorder"
)

// FileStorage presents storage on file
type FileStorage struct {
	m map[string]string
	// user urls grouped by uid
	muser    map[string][]models.ShrURL
	mtx      sync.RWMutex
	fileName string
	rec      *recorder.Recorder
	index    int
}

// NewFileStorage is created new storage on file
func NewFileStorage(filename string) *FileStorage {
	newRec, err := recorder.NewRecorder()
	if err != nil {
		return nil
	}

	return &FileStorage{
		m:        make(map[string]string),
		muser:    make(map[string][]models.ShrURL),
		fileName: filename,
		rec:      newRec,
	}
}

// LoadFromFile is load storage from file
func (fs *FileStorage) LoadFromFile() error {
	fs.mtx.Lock()
	defer fs.mtx.Unlock()

	file, err := os.OpenFile(fs.fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	fs.m, err = fs.rec.ReadAllRecords(file)
	if err != nil {
		return err
	}

	fs.index = len(fs.m)

	return nil
}

// StoreURLCtx add url alias and original url to storage
func (fs *FileStorage) StoreURLCtx(ctx context.Context, url models.ShrURL) error {
	fs.mtx.Lock()
	defer fs.mtx.Unlock()

	if fs.isURLExist(url.URL) {
		// url exist
		return ErrURLExists
	}

	// put url
	fs.m[url.Alias] = url.URL
	// put user url
	fs.muser[url.UserID] = append(fs.muser[url.UserID], url)

	file, err := os.OpenFile(fs.fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	fs.index++

	nr := recorder.Record{
		UUID:        strconv.Itoa(fs.index),
		ShortURL:    url.Alias,
		OriginalURL: url.URL,
	}

	if err := fs.rec.WriteRecord(file, &nr); err != nil {
		return err
	}

	return nil
}

// StoreBatchURLCtx stores batch urls
func (fs *FileStorage) StoreBatchURLCtx(ctx context.Context, urls []models.ShrURL) error {
	fs.mtx.Lock()
	defer fs.mtx.Unlock()

	file, err := os.OpenFile(fs.fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	for _, url := range urls {
		if fs.isURLExist(url.URL) {
			// url exist
			continue
		}

		// put url
		fs.m[url.Alias] = url.URL
		// put user url
		fs.muser[url.UserID] = append(fs.muser[url.UserID], url)

		fs.index++

		nr := recorder.Record{
			UUID:        strconv.Itoa(fs.index),
			ShortURL:    url.Alias,
			OriginalURL: url.URL,
		}

		if err := fs.rec.WriteRecord(file, &nr); err != nil {
			return err
		}
	}

	return nil
}

// GetURLCtx returns url alias and original url by alias
// if alias is not exist return an error
func (fs *FileStorage) GetURLCtx(ctx context.Context, alias string) (models.ShrURL, error) {
	fs.mtx.RLock()
	defer fs.mtx.RUnlock()

	url, ok := fs.m[alias]
	if !ok {
		return models.ShrURL{}, ErrURLNotFound
	}
	return models.ShrURL{Alias: alias, URL: url}, nil
}

// GetAliasCtx returns stored alias by url
// if alias is not exist return an error
func (fs *FileStorage) GetAliasCtx(ctx context.Context, url string) (models.ShrURL, error) {
	fs.mtx.RLock()
	defer fs.mtx.RUnlock()
	for k, v := range fs.m {
		if strings.Compare(v, url) == 0 {
			return models.ShrURL{Alias: k, URL: v}, nil
		}
	}
	return models.ShrURL{}, ErrAliasNotFound
}

// GetUserURLsCtx returns all user URLs by user ID
func (fs *FileStorage) GetUserURLsCtx(ctx context.Context, userID string) ([]models.ShrURL, error) {
	fs.mtx.RLock()
	defer fs.mtx.RUnlock()
	urls, ok := fs.muser[userID]
	if !ok {
		return nil, ErrUserNotFound
	}
	return urls, nil
}

// DeleteUserURLsCtx deletes user URLs
func (fs *FileStorage) DeleteUserURLsCtx(ctx context.Context, userID string, aliases []string) error {
	// TODO
	return nil
}

// isURLExist checks existing url
func (fs *FileStorage) isURLExist(url string) bool {
	// does the url exist?
	for _, v := range fs.m {
		if strings.Compare(v, url) == 0 {
			// url exist
			return true
		}
	}
	return false
}

// GetUserCountCtx returns user count
func (fs *FileStorage) GetUserCountCtx(ctx context.Context) (int, error) {
	return len(fs.muser), nil
}

// GetURLCountCtx returns the number of shortened urls
func (fs *FileStorage) GetURLCountCtx(ctx context.Context) (int, error) {
	return len(fs.m), nil
}
