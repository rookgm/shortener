package grpc

import (
	"context"
	"github.com/rookgm/shortener/internal/client"
	"github.com/rookgm/shortener/internal/grpc/pb"
	"github.com/rookgm/shortener/internal/models"
	"github.com/rookgm/shortener/internal/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
	"net"
	"testing"
)

func createTestServer(t *testing.T) (pb.ShortenerClient, func()) {
	store := storage.NewMemStorage()

	auth := client.NewAuthToken([]byte("secretkey"))

	ch := make(chan models.UserDeleteTask, 1)

	listen := bufconn.Listen(1024)

	server := grpc.NewServer(grpc.UnaryInterceptor(AuthInterceptor(auth)))
	shServer := NewShortenerServer(
		store,
		"http://localhost:8080/",
		ch,
		"")

	pb.RegisterShortenerServer(server, shServer)

	go func() {
		if err := server.Serve(listen); err != nil {
			t.Fatal("error starting server")
		}
	}()

	dialer := func(context.Context, string) (net.Conn, error) {
		return listen.Dial()
	}

	conn, err := grpc.NewClient("passthrough://bufnet",
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)

	client := pb.NewShortenerClient(conn)

	cleanup := func() {
		conn.Close()
		server.Stop()
		listen.Close()
	}

	return client, cleanup
}

func TestShortenerServer_ShortenURL(t *testing.T) {
	client, cleanup := createTestServer(t)
	defer cleanup()

	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{
			name:    "valid_url",
			url:     "https://practicum.yandex.ru/",
			wantErr: false,
		},
		{
			name:    "empty_url",
			url:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := metadata.AppendToOutgoingContext(context.Background(), "auth_token", "test-token")

			resp, err := client.ShortenURL(ctx, &pb.ShortenURLRequest{
				Url: tt.url,
			})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, resp.ShortUrl)
				assert.Contains(t, resp.ShortUrl, "http://localhost:8080")
			}
		})
	}
}

func TestShortenerServer_BatchURL(t *testing.T) {
	client, cleanup := createTestServer(t)
	defer cleanup()

	ctx := metadata.AppendToOutgoingContext(context.Background(), "auth_token", "test-token")

	req := &pb.BatchURLRequest{
		OrigUrls: []*pb.BatchURLItem{
			{
				CorrelationId: "1",
				Url:           "https://practicum.yandex.ru/1",
			},
			{
				CorrelationId: "2",
				Url:           "https://practicum.yandex.ru/2",
			},
		},
	}

	resp, err := client.BatchURL(ctx, req)
	require.NoError(t, err)
	assert.Len(t, resp.ShortUrls, 2)

	for i, item := range resp.ShortUrls {
		assert.Equal(t, req.OrigUrls[i].CorrelationId, item.CorrelationId)
		assert.Contains(t, item.Url, "http://localhost:8080")
	}
}

func TestShortenerServer_GetUserURL(t *testing.T) {
	client, cleanup := createTestServer(t)
	defer cleanup()

	ctx := metadata.AppendToOutgoingContext(context.Background(), "auth_token", "test-token")

	urls := []string{"https://practicum.yandex.ru/1", "https://practicum.yandex.ru/2"}
	for _, url := range urls {
		_, err := client.ShortenURL(ctx, &pb.ShortenURLRequest{
			Url: url,
		})
		require.NoError(t, err)
	}

	resp, err := client.GetUserURL(ctx, &pb.GetUserURLRequest{})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(resp.Urls), 2)
}

func TestShortenerServer_Ping(t *testing.T) {
	client, cleanup := createTestServer(t)
	defer cleanup()

	_, err := client.Ping(context.Background(), &pb.PingRequest{})
	assert.NoError(t, err)
}
