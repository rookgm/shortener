package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/rookgm/shortener/internal/client"
	"github.com/rookgm/shortener/internal/models"
	"github.com/rookgm/shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetUserUrlsHandler(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		setup          func(t *testing.T) *storage.MockURLStorage
		wantStatusCode int
		wantBody       []UserURL
	}{
		{
			name: "valid_request_return_200",
			setup: func(t *testing.T) *storage.MockURLStorage {

				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				storeMock := storage.NewMockURLStorage(ctrl)
				storeMock.EXPECT().GetUserURLsCtx(gomock.Any(), gomock.Any()).Return([]models.ShrURL{
					{Alias: "5LBgy9",
						URL:     "http://uv4nq5mt9qkh7z.ru",
						UserID:  "c81514ed-b47a-4d39-9591-b904db48a07a",
						Deleted: false},
				}, nil).AnyTimes()
				return storeMock
			},
			wantStatusCode: http.StatusOK,
			wantBody: []UserURL{
				{
					ShortURL:    "http://localhost/5LBgy9",
					OriginalURL: "http://uv4nq5mt9qkh7z.ru",
				},
			},
		},
		{
			name: "no_content_return_204",
			body: "",
			setup: func(t *testing.T) *storage.MockURLStorage {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				storeMock := storage.NewMockURLStorage(ctrl)
				storeMock.EXPECT().GetUserURLsCtx(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
				return storeMock
			},
			wantStatusCode: http.StatusNoContent,
			wantBody:       nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/api/user/urls", bytes.NewBufferString(tt.body))
			if err != nil {
				t.Fatal("cannot create request", zap.Error(err))
			}
			w := httptest.NewRecorder()
			auth := client.NewAuthToken([]byte("secretkey"))

			st := tt.setup(t)

			handler := GetUserUrlsHandler(st, "http://localhost", auth)
			handler(w, req)

			res := w.Result()
			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			var got []UserURL

			json.Unmarshal(resBody, &got)

			if diff := cmp.Diff(tt.wantBody, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}

		})
	}
}
