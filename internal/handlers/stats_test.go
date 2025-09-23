package handlers

import (
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/rookgm/shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatsHandler(t *testing.T) {
	type args struct {
		store storage.URLStorage
	}
	tests := []struct {
		name           string
		setup          func(t *testing.T) *storage.MockURLStorage
		wantStatusCode int
		wantBody       StatsResponse
	}{
		{
			name: "valid_request_return_200",
			setup: func(t *testing.T) *storage.MockURLStorage {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				storeMock := storage.NewMockURLStorage(ctrl)
				// user count
				storeMock.EXPECT().GetUserCountCtx(gomock.Any()).Return(1, nil).AnyTimes()
				// url count
				storeMock.EXPECT().GetURLCountCtx(gomock.Any()).Return(1, nil).AnyTimes()

				return storeMock
			},
			wantStatusCode: http.StatusOK,
			wantBody: StatsResponse{
				Urls:  1,
				Users: 1,
			},
		},
		{
			name: "internal_error_return_500",
			setup: func(t *testing.T) *storage.MockURLStorage {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				storeMock := storage.NewMockURLStorage(ctrl)
				// user count
				storeMock.EXPECT().GetUserCountCtx(gomock.Any()).Return(0, errors.New("internal error")).AnyTimes()
				// url count
				storeMock.EXPECT().GetURLCountCtx(gomock.Any()).Return(0, errors.New("internal error")).AnyTimes()

				return storeMock
			},
			wantStatusCode: http.StatusInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, "/api/internal/stats", nil)
			if err != nil {
				t.Fatal("cannot create request", zap.Error(err))
			}
			w := httptest.NewRecorder()

			st := tt.setup(t)

			handler := StatsHandler(st)
			handler(w, req)

			res := w.Result()
			assert.Equal(t, tt.wantStatusCode, res.StatusCode)
			resBody, err := io.ReadAll(res.Body)
			defer res.Body.Close()
			require.NoError(t, err)

			var got StatsResponse

			json.Unmarshal(resBody, &got)

			if diff := cmp.Diff(tt.wantBody, got); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
