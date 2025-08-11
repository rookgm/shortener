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
	fmt.Println(res.StatusCode)
	defer res.Body.Close()

	// Output:
	// 307
}

func ExamplePostHandler() {
	st := storage.NewMemStorage()
	auth := client.NewAuthToken([]byte("secretkey"))
	handler := PostHandler(st, "http://localhost:8080", auth)
	originalURL := "http://practicum.yandex.ru/test"

	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(originalURL))
	w := httptest.NewRecorder()

	handler(w, request)

	res := w.Result()
	fmt.Println(res.StatusCode)
	defer res.Body.Close()

	// Output:
	// 201
}

func ExampleAPIShortenHandler() {
	st := storage.NewMemStorage()
	handler := APIShortenHandler(st, "http://localhost:8080")
	orig := APIShortenReq{URL: "https://practicum.yandex.ru/"}
	body, _ := json.Marshal(orig)
	request := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler(w, request)

	res := w.Result()
	fmt.Println(res.StatusCode)
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var resp APIShortenResp

	if err := json.Unmarshal(resBody, &resp); err != nil {
		log.Fatal(err)
	}

	// Output:
	// 201
}

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

	handler := GetUserUrlsHandler(st, "http://localhost", auth)
	handler(w, req)

	res := w.Result()
	fmt.Println(res.StatusCode)
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var got []UserURL

	if err := json.Unmarshal(resBody, &got); err != nil {
		log.Fatal(err)
	}

	// Output:
	// 200
}
