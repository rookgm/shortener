package handlers

import (
	"github.com/rookgm/shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestStatsHandler(t *testing.T) {
	type args struct {
		store storage.URLStorage
	}
	tests := []struct {
		name string
		args args
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, StatsHandler(tt.args.store), "StatsHandler(%v)", tt.args.store)
		})
	}
}
