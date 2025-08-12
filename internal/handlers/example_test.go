package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/rookgm/shortener/internal/client"
	"github.com/rookgm/shortener/internal/models"
	"github.com/rookgm/shortener/internal/storage"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
)

// ExampleGetHandler is example get original URL by shortened URL
func ExampleGetHandler() {

	st := storage.NewMemStorage()

	url := models.ShrURL{
		Alias: "EwHXdJfB",
		URL:   "https://practicum.yandex.ru/",
	}

	err := st.StoreURLCtx(context.Background(), url)
	if err != nil {
		log.Fatal(err)
	}

	router := chi.NewRouter()
	router.Get("/{id}", GetHandler(st))

	request := httptest.NewRequest(http.MethodGet, "/"+url.Alias, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, request)

	res := w.Result()
	fmt.Printf("StatusCode: %d\n", res.StatusCode)
	defer res.Body.Close()
	fmt.Printf("Location: %s\n", res.Header.Get("Location"))

	// Output:
	// StatusCode: 307
	// Location: https://practicum.yandex.ru/
}

// ExamplePostHandler is example getting a shortened URL
func ExamplePostHandler() {

	st := storage.NewMemStorage()

	st.StoreURLCtx(context.Background(), models.ShrURL{
		Alias: "EwHXdJfB",
		URL:   "http://practicum.yandex.ru/test",
	})

	auth := client.NewAuthToken([]byte("secretkey"))
	handler := PostHandler(st, "http://localhost:8080", auth)
	originalURL := "http://practicum.yandex.ru/test"

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	w := httptest.NewRecorder()

	handler(w, request)

	res := w.Result()
	fmt.Printf("StatusCode: %d\n", res.StatusCode)
	defer res.Body.Close()
	out, _ := io.ReadAll(res.Body)
	fmt.Printf("Response: %s\n", string(out))

	// Output:
	// StatusCode: 409
	// Response: http://localhost:8080/EwHXdJfB
}

// ExampleAPIShortenHandler is example shortening URL in JSON format
func ExampleAPIShortenHandler() {
	st := storage.NewMemStorage()
	st.StoreURLCtx(context.Background(), models.ShrURL{
		Alias: "EwHXdJfB",
		URL:   "https://practicum.yandex.ru/test",
	})
	handler := APIShortenHandler(st, "http://localhost:8080")
	orig := APIShortenReq{URL: "https://practicum.yandex.ru/test"}
	body, _ := json.Marshal(orig)
	request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler(w, request)

	res := w.Result()
	fmt.Printf("StatusCode: %d\n", res.StatusCode)
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response: %s\n", string(resBody))

	var resp APIShortenResp

	if err := json.Unmarshal(resBody, &resp); err != nil {
		log.Fatal(err)
	}

	// Output:
	// StatusCode: 409
	// Response: {"result":"http://localhost:8080/EwHXdJfB"}
}

// ExampleGetUserUrlsHandler is receiving all user urls
func ExampleGetUserUrlsHandler() {
	st := storage.NewMemStorage()
	err := st.StoreURLCtx(context.Background(), models.ShrURL{
		Alias:   "5LBgy9",
		URL:     "http://uv4nq5mt9qkh7z.ru",
		Deleted: false,
	})
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodGet, "/api/user/urls", nil)
	if err != nil {
		log.Fatal(err)
	}
	w := httptest.NewRecorder()
	auth := client.NewAuthToken([]byte("secretkey"))

	handler := GetUserUrlsHandler(st, "http://localhost:8080", auth)
	handler(w, req)

	res := w.Result()
	fmt.Printf("StatusCode: %d\n", res.StatusCode)
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Response: %s\n", string(resBody))

	var got []UserURL

	if err := json.Unmarshal(resBody, &got); err != nil {
		log.Fatal(err)
	}

	fmt.Println()

	// Output:
	// StatusCode: 200
	// Response: [{"short_url":"http://localhost:8080/5LBgy9","original_url":"http://uv4nq5mt9qkh7z.ru"}]
}
