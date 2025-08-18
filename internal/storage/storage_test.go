package storage

import (
	"context"
	"os"
	"testing"

	"github.com/rookgm/shortener/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestStorage_SetAndGet(t *testing.T) {

	fileName := "storage_test.json"
	defer os.Remove(fileName)

	st := NewFileStorage(fileName)
	assert.NotEqual(t, st, nil, "storage is nil")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// set1
	url1 := models.ShrURL{
		Alias: "4rSPg8ap",
		URL:   "http://yandex.ru",
	}
	err := st.StoreURLCtx(ctx, url1)
	assert.NoError(t, err, "set")

	// set 2
	url2 := models.ShrURL{
		Alias: "edVPg3ks",
		URL:   "http://ya.ru",
	}
	err = st.StoreURLCtx(ctx, url2)
	assert.NoError(t, err, "set")

	// set 3
	url3 := models.ShrURL{
		Alias: "dG56Hqxm",
		URL:   "http://practicum.yandex.ru",
	}
	err = st.StoreURLCtx(ctx, url3)
	assert.NoError(t, err, "set")

	v, err := st.GetURLCtx(ctx, url1.Alias)
	assert.NoError(t, err, "get")
	assert.Equal(t, url1.URL, v.URL)

	v, err = st.GetURLCtx(ctx, url2.Alias)
	assert.NoError(t, err, "get")
	assert.Equal(t, url2.URL, v.URL)

	v, err = st.GetURLCtx(ctx, url3.Alias)
	assert.NoError(t, err, "get")
	assert.Equal(t, url3.URL, v.URL)

	// file storage
	fst := NewFileStorage(fileName)
	assert.NotEqual(t, st, nil, "storage is nil")

	err = fst.LoadFromFile()
	assert.NoError(t, err, "LoadFromFile")

	v, err = fst.GetURLCtx(ctx, url1.Alias)
	assert.NoError(t, err, "get")
	assert.Equal(t, url1.URL, v.URL)

	v, err = fst.GetURLCtx(ctx, url2.Alias)
	assert.NoError(t, err, "get")
	assert.Equal(t, url2.URL, v.URL)

	v, err = fst.GetURLCtx(ctx, url3.Alias)
	assert.NoError(t, err, "get")
	assert.Equal(t, url3.URL, v.URL)
}
