package storage

import (
	"github.com/rookgm/shortener/internal/models"
	"github.com/rookgm/shortener/internal/recorder"
	"os"
	"strconv"
	"sync"
)

// FileStorage presents storage on file
type FileStorage struct {
	m        map[string]string
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
		fileName: filename,
		rec:      newRec,
	}
}

// LoadFromFile is load storage from file
func (st *FileStorage) LoadFromFile() error {
	st.mtx.Lock()
	defer st.mtx.Unlock()

	file, err := os.OpenFile(st.fileName, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	st.m, err = st.rec.ReadAllRecords(file)
	if err != nil {
		return err
	}

	return nil
}

// StoreURL add url alias and original url to storage
func (st *FileStorage) StoreURL(url models.ShrURL) error {
	st.mtx.Lock()
	defer st.mtx.Unlock()

	st.m[url.Alias] = url.URL

	file, err := os.OpenFile(st.fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	st.index++

	nr := recorder.Record{
		UUID:        strconv.Itoa(st.index),
		ShortURL:    url.Alias,
		OriginalURL: url.URL,
	}

	if err := st.rec.WriteRecord(file, &nr); err != nil {
		return err
	}

	return nil
}

// GetURL returns url alias and original url by alias
// if alias is not exist return an error
func (st *FileStorage) GetURL(alias string) (models.ShrURL, error) {
	st.mtx.RLock()
	defer st.mtx.RUnlock()

	url, ok := st.m[alias]
	if !ok {
		return models.ShrURL{}, ErrURLNotFound
	}
	return models.ShrURL{Alias: alias, URL: url}, nil
}
