package server

import (
	"context"
	"encoding/hex"
	"errors"
	grpcserver "github.com/rookgm/shortener/internal/grpc"
	"github.com/rookgm/shortener/internal/grpc/pb"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rookgm/shortener/config"
	"github.com/rookgm/shortener/internal/client"
	"github.com/rookgm/shortener/internal/db"
	"github.com/rookgm/shortener/internal/handlers"
	"github.com/rookgm/shortener/internal/logger"
	"github.com/rookgm/shortener/internal/middleware"
	"github.com/rookgm/shortener/internal/models"
	"github.com/rookgm/shortener/internal/storage"
	"go.uber.org/zap"
)

const authTokenKey = "f53ac685bbceebd75043e6be2e06ee07"

const (
	serverCertFileName = "cert/server.crt"
	serverKeyFileName  = "cert/server.key"
)

// Run is prepare server and runs it. Choose type of storage and create it.
// Launch delete worker.Setup main routers.
func Run(config *config.Config) error {
	if config == nil {
		return errors.New("config is nil")
	}
	// if https is enabled, then loads cert
	if config.EnableHTTPS {
		// check existing server's key files
		if _, err := os.Stat(serverCertFileName); errors.Is(err, os.ErrNotExist) {
			logger.Log.Error("server cert file is not exist")
			return err
		}
		if _, err := os.Stat(serverKeyFileName); errors.Is(err, os.ErrNotExist) {
			logger.Log.Error("server key file is not exist")
			return err
		}
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	defer stop()

	var sdb *db.DataBase
	var st storage.URLStorage
	var err error

	// detect type of storage
	if config.DataBaseDSN != "" {
		// open shortener db
		sdb, err = db.OpenCtx(ctx, config.DataBaseDSN)
		if err != nil {
			return errors.New("can open db")
		}
		defer sdb.Close()
		// create db storage
		st, err = storage.NewDBStorage(sdb)
		if err != nil {
			return err
		}
	} else if config.StoragePath != "" {
		// create file storage
		st = storage.NewFileStorage(config.StoragePath)
		// load storage from file
		if err := st.LoadFromFile(); err != nil {
			return err
		}
	} else {
		// create storage on memory
		st = storage.NewMemStorage()
	}

	key, err := hex.DecodeString(authTokenKey)
	if err != nil {
		logger.Log.Error("can not extract key")
		return err
	}

	// faInCh channel accepts data for batch alias deletion
	fanInCh := make(chan models.UserDeleteTask, 1000)

	// run delete worker
	var batch []models.UserDeleteTask
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				if len(batch) > 0 {
					for _, b := range batch {
						logger.Log.Debug("delete user urls", zap.String("uid", b.UID))
						if err := st.DeleteUserURLsCtx(ctx, b.UID, b.Aliases); err != nil {
							logger.Log.Error("can't delete user urls", zap.String("uid", b.UID), zap.Error(err))
						}
					}
				}
				logger.Log.Debug("delete worker is stopped")
				return
			case toDelete, ok := <-fanInCh:
				if !ok {
					if len(batch) > 0 {
						if err := st.DeleteUserURLsCtx(ctx, toDelete.UID, toDelete.Aliases); err != nil {
							logger.Log.Error("can't delete user urls", zap.String("uid", toDelete.UID), zap.Error(err))
						}
					}
				}
				logger.Log.Debug("got delete task in fanInCh", zap.Any("task", toDelete))
				batch = append(batch, toDelete)
			case <-ticker.C:
				logger.Log.Debug("ticker tick")
				logger.Log.Debug("starting delete user urls...")
				for _, b := range batch {
					logger.Log.Debug("delete user urls", zap.String("uid", b.UID))
					if err := st.DeleteUserURLsCtx(ctx, b.UID, b.Aliases); err != nil {
						logger.Log.Error("can't delete user urls", zap.String("uid", b.UID), zap.Error(err))
					}
				}
				batch = nil
			}
			logger.Log.Debug("finished delete user urls")
		}
	}()

	token := client.NewAuthToken(key)

	router := chi.NewRouter()
	router.Use(logger.Middleware)
	router.Use(middleware.GzipMiddleware)
	router.Use(func(next http.Handler) http.Handler {
		return middleware.Auth(token, next)
	})

	router.Route("/", func(r chi.Router) {
		router.Post("/", handlers.PostHandler(st, config.BaseURL, token))
		router.Get("/{id}", handlers.GetHandler(st))
		router.Post("/api/shorten", handlers.APIShortenHandler(st, config.BaseURL))
		router.Get("/ping", handlers.PingHandler(sdb))
		router.Post("/api/shorten/batch", handlers.PostBatchHandler(st, config.BaseURL))
		router.Get("/api/user/urls", handlers.GetUserUrlsHandler(st, config.BaseURL, token))
		router.Delete("/api/user/urls", handlers.DeleteUserUrlsHandler(st, token, fanInCh))

		if config.DebugMode {
			r.HandleFunc("/debug/pprof/*", pprof.Index)
			r.Get("/debug/pprof/profile", pprof.Profile)
		}

		r.Route("/api/internal", func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return middleware.CheckTrustedSubNet(config.TrustedSubNet, next)
			})
			r.Get("/stats", handlers.StatsHandler(st))
		})
	})

	// wait group for http and grpc servers
	var wg sync.WaitGroup
	srvErrCh := make(chan error, 2)

	// set server parameters
	srv := http.Server{
		Addr:    config.ServerAddr,
		Handler: router,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		// run server supporting https connections
		if config.EnableHTTPS {
			logger.Log.Info("Starting HTTPs server", zap.String("address", config.ServerAddr))
			if err := http.ListenAndServeTLS(config.ServerAddr, serverCertFileName, serverKeyFileName, router); err != nil && !errors.Is(err, http.ErrServerClosed) {
				logger.Log.Fatal("Error starting https server", zap.Error(err))
				srvErrCh <- err
			}
		}
		// run server with http
		logger.Log.Info("Starting HTTP server", zap.String("address", config.ServerAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatal("Error starting http server", zap.Error(err))
			srvErrCh <- err
		}
	}()

	var grpcListen net.Listener
	var grpcServer *grpc.Server

	// if grpc server is enabled
	if config.EnableGRPC {
		// prepare grpc server
		grpcListen, err = net.Listen("tcp", config.GRPCSeverAddr)
		if err != nil {
			logger.Log.Fatal("Error listen on", zap.String("address", config.GRPCSeverAddr), zap.Error(err))
		}
		// create grpc server
		grpcServer = grpc.NewServer(grpc.ChainUnaryInterceptor(
			grpcserver.LogInterceptor(logger.Log),
			grpcserver.AuthInterceptor(token),
		))
		shServer := grpcserver.NewShortenerServer(st, config.BaseURL, fanInCh, config.TrustedSubNet)
		// register shortener server
		pb.RegisterShortenerServer(grpcServer, shServer)

		wg.Add(1)
		go func() {
			defer wg.Done()
			// run grpc server
			logger.Log.Info("Starting gRPC server", zap.String("address", config.GRPCSeverAddr))
			if err := grpcServer.Serve(grpcListen); err != nil {
				logger.Log.Fatal("Error starting grpc server", zap.Error(err))
			}
		}()
	}

	select {
	case <-ctx.Done():
		logger.Log.Info("server is done")
	case err := <-srvErrCh:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	<-shutdownCtx.Done()

	// shutdown http server
	logger.Log.Info("Shutting http server")
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Log.Error("Error shutdown server", zap.Error(err))
	}
	// shutdown grpc server
	logger.Log.Info("Shutting down grpc server")
	if config.EnableGRPC && grpcServer != nil {
		grpcServer.GracefulStop()
	}

	// close storage
	if sdb != nil {
		// close db
		if err := sdb.DB.Close(); err != nil {
			logger.Log.Error("Error closing database", zap.Error(err))
		}
		sdb.Close()
	}

	logger.Log.Info("server is finished")

	return nil
}
