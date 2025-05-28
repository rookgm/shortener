package storage

import (
	"github.com/rookgm/shortener/internal/recorder"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestStorage_SetAndGet(t *testing.T) {

	rec1 := recorder.Record{
		ShortURL:    "4rSPg8ap",
		OriginalURL: "http://yandex.ru",
	}

	rec2 := recorder.Record{
		ShortURL:    "edVPg3ks",
		OriginalURL: "http://ya.ru",
	}

	rec3 := recorder.Record{
		ShortURL:    "dG56Hqxm",
		OriginalURL: "http://practicum.yandex.ru",
	}

	fileName := "storage_test.json"
	defer os.Remove(fileName)

	st := NewStorage(fileName)
	assert.NotEqual(t, st, nil, "storage is nil")

	err := st.Set(rec1.ShortURL, rec1.OriginalURL)
	assert.NoError(t, err, "set")

	err = st.Set(rec2.ShortURL, rec2.OriginalURL)
	assert.NoError(t, err, "set")

	err = st.Set(rec3.ShortURL, rec3.OriginalURL)
	assert.NoError(t, err, "set")

	v, err := st.Get(rec1.ShortURL)
	assert.NoError(t, err, "get")
	assert.Equal(t, rec1.OriginalURL, v)

	v, err = st.Get(rec2.ShortURL)
	assert.NoError(t, err, "get")
	assert.Equal(t, rec2.OriginalURL, v)

	v, err = st.Get(rec3.ShortURL)
	assert.NoError(t, err, "get")
	assert.Equal(t, rec3.OriginalURL, v)

	fst := NewStorage(fileName)
	assert.NotEqual(t, st, nil, "storage is nil")

	err = fst.LoadFromFile()
	assert.NoError(t, err, "LoadFromFile")

	v, err = fst.Get(rec1.ShortURL)
	assert.NoError(t, err, "get")
	assert.Equal(t, rec1.OriginalURL, v)

	v, err = fst.Get(rec2.ShortURL)
	assert.NoError(t, err, "get")
	assert.Equal(t, rec2.OriginalURL, v)

	v, err = fst.Get(rec3.ShortURL)
	assert.NoError(t, err, "get")
	assert.Equal(t, rec3.OriginalURL, v)
}
