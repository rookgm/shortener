package storage

import (
	"errors"
	"github.com/rookgm/shortener/internal/recorder"
	"os"
	"strconv"
	"sync"
)

type Storage struct {
	m        map[string]string
	mtx      sync.RWMutex
	fileName string
	rec      *recorder.Recorder
	index    int
}

// NewStorage is created new storage
func NewStorage(filename string) *Storage {
	newRec, err := recorder.NewRecorder()
	if err != nil {
		return nil
	}

	return &Storage{
		m:        make(map[string]string),
		fileName: filename,
		rec:      newRec,
	}
}

func (st *Storage) LoadFromFile() error {
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

// Set add a new key value pair to storage
func (st *Storage) Set(key, value string) error {
	st.mtx.Lock()
	defer st.mtx.Unlock()

	st.m[key] = value

	file, err := os.OpenFile(st.fileName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	st.index++

	nr := recorder.Record{
		UUID:        strconv.Itoa(st.index),
		ShortURL:    key,
		OriginalURL: value,
	}

	if err := st.rec.WriteRecord(file, &nr); err != nil {
		return err
	}

	return nil
}

// Get returns the value of a key
// if key not exists return an error
func (st *Storage) Get(key string) (string, error) {
	st.mtx.RLock()
	defer st.mtx.RUnlock()

	value, ok := st.m[key]
	if !ok {
		return "", errors.New("key not exists")
	}
	return value, nil
}
