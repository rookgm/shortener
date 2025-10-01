package grpc

import (
	"context"
	"github.com/rookgm/shortener/internal/client"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"time"
)

const authToken = "auth_token"

// AuthInterceptor is a unary server interceptor for authentication
func AuthInterceptor(tokenService client.AuthToken) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if info.FullMethod == "/shortener.Shortener/Ping" {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}
		// get auth token
		tokens := md.Get(authToken)
		if len(tokens) == 0 {
			// generate new token
			token, err := tokenService.Create()
			if err != nil {
				return nil, err
			}
			header := metadata.Pairs(authToken, token)
			grpc.SendHeader(ctx, header)
			md.Append(authToken, token)
			ctx = metadata.NewIncomingContext(ctx, md)
		}

		return handler(ctx, req)
	}
}

// LogInterceptor is a unary server interceptor for logging
func LogInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		t := time.Now()

		resp, err := handler(ctx, req)
		dt := time.Since(t)

		if err != nil {
			logger.Error("gRPC request failed",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", dt),
				zap.Error(err),
			)
		} else {
			logger.Info("gRPC request is completed",
				zap.String("method", info.FullMethod),
				zap.Duration("duration", dt),
			)
		}

		return resp, err
	}
}
