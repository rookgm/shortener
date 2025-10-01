package grpc

import (
	"context"
	"errors"
	"github.com/rookgm/shortener/internal/grpc/pb"
	"github.com/rookgm/shortener/internal/logger"
	"github.com/rookgm/shortener/internal/models"
	"github.com/rookgm/shortener/internal/random"
	"github.com/rookgm/shortener/internal/storage"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net"
	"net/url"
)

// ShortenerServer implements ShortenerServer interface
type ShortenerServer struct {
	pb.UnimplementedShortenerServer
	// url storage
	storage storage.URLStorage
	// authority
	baseURL string
	// user's url for asynchronous deletion
	urlsToDeleteCh chan<- models.UserDeleteTask
	// trusted subnet
	trustedNet *net.IPNet
}

// NewShortenerServer creates new shortener grpc server
func NewShortenerServer(storage storage.URLStorage, baseURL string, urlsToDeleteCh chan<- models.UserDeleteTask, trustedSubNet string) *ShortenerServer {
	// parse CIDR
	_, inet, _ := net.ParseCIDR(trustedSubNet)

	return &ShortenerServer{
		storage:        storage,
		baseURL:        baseURL,
		urlsToDeleteCh: urlsToDeleteCh,
		trustedNet:     inet,
	}
}

// getUserIDFromContext user id from context metadata
func (ss *ShortenerServer) getUserIDFromContext(ctx context.Context) (string, error) {
	// get metadata from context
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.InvalidArgument, "can not get metadata")
	}
	// extract user id
	userIDs := md.Get("auth_token")
	if len(userIDs) == 0 {
		return "", status.Error(codes.Unauthenticated, "userid is not found in metadata")
	}

	return userIDs[0], nil
}

// getClientIPFromContext extracts client ip from context metadata
func (ss *ShortenerServer) getClientIPFromContext(ctx context.Context) (net.IP, error) {
	// get client ip
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.InvalidArgument, "can not get metadata")
	}

	// get client ip from request header value
	v := md.Get("x-real-ip")
	if len(v) == 0 {
		return nil, status.Error(codes.InvalidArgument, "client ip is missed")
	}
	// try parse ip
	clientIP := net.ParseIP(v[0])
	if clientIP == nil {
		return nil, status.Error(codes.InvalidArgument, "parse ip is failed")
	}

	return clientIP, nil
}

// ShortenURL accepts URL and returns shortened URL
func (ss *ShortenerServer) ShortenURL(ctx context.Context, req *pb.ShortenURLRequest) (*pb.ShortenURLResponse, error) {
	// try get user id from context
	userID, err := ss.getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userid is missing")
	}

	// generate shorten url
	iurl := models.ShrURL{
		Alias:  random.RandString(6),
		URL:    req.Url,
		UserID: userID,
	}

	// put it storage
	if err := ss.storage.StoreURLCtx(ctx, iurl); err != nil {
		if errors.Is(err, storage.ErrURLExists) {
			ourl, err := ss.storage.GetAliasCtx(ctx, iurl.URL)
			if err != nil {
				return nil, status.Error(codes.Internal, "internal error")
			}
			rurl, _ := url.JoinPath(ss.baseURL, ourl.Alias)
			return &pb.ShortenURLResponse{ShortUrl: rurl}, status.Error(codes.AlreadyExists, "URL is already exist")
		} else {
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	rurl, err := url.JoinPath(ss.baseURL, iurl.Alias)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.ShortenURLResponse{ShortUrl: rurl}, nil
}

// ShortenURLJSON accepts URL and returns shortened URL in json format
func (ss *ShortenerServer) ShortenURLJSON(ctx context.Context, req *pb.ShortenURLJSONRequest) (*pb.ShortenURLJSONResponse, error) {
	// try to get user id from context
	userID, err := ss.getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userid is missing")
	}

	// generate shorten url
	iurl := models.ShrURL{
		Alias:  random.RandString(6),
		URL:    req.Url,
		UserID: userID,
	}

	// put it storage
	if err := ss.storage.StoreURLCtx(ctx, iurl); err != nil {
		if errors.Is(err, storage.ErrURLExists) {
			ourl, err := ss.storage.GetAliasCtx(ctx, iurl.URL)
			if err != nil {
				return nil, status.Error(codes.Internal, "internal error")
			}
			rurl, _ := url.JoinPath(ss.baseURL, ourl.Alias)
			return &pb.ShortenURLJSONResponse{ShortUrl: rurl}, status.Error(codes.AlreadyExists, "URL is already exist")
		} else {
			return nil, status.Error(codes.Internal, "internal error")
		}
	}

	rurl, err := url.JoinPath(ss.baseURL, iurl.Alias)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.ShortenURLJSONResponse{ShortUrl: rurl}, nil
}

// GetOriginalURL gets original URL by shortened URL
func (ss *ShortenerServer) GetOriginalURL(ctx context.Context, req *pb.GetOriginalURLRequest) (*pb.GetOriginalURLResponse, error) {
	rurl, err := ss.storage.GetURLCtx(ctx, req.ShortUrl)
	if err != nil {
		if errors.Is(err, storage.ErrURLNotFound) {
			return nil, status.Error(codes.NotFound, "url is not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.GetOriginalURLResponse{
		OriginalUrl: rurl.URL,
	}, nil
}

// BatchURL processes of original URLs and returns shortened URLs
func (ss *ShortenerServer) BatchURL(ctx context.Context, req *pb.BatchURLRequest) (*pb.BatchURLResponse, error) {
	if len(req.OrigUrls) == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid argument")
	}

	var batchURL []models.ShrURL

	// prepare urls before batch processing
	for _, item := range req.OrigUrls {
		// get shorten url
		iurl := models.ShrURL{
			Alias: random.RandString(6),
			URL:   item.Url,
		}
		batchURL = append(batchURL, iurl)
	}
	// store batched url
	if err := ss.storage.StoreBatchURLCtx(ctx, batchURL); err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	// batch response
	batchResp := make([]*pb.BatchURLItem, len(batchURL))

	// forming batch result
	for i, breq := range req.OrigUrls {
		// get shorten url by original url
		rurl, err := ss.storage.GetAliasCtx(ctx, breq.Url)
		if err != nil {
			continue
		}
		// make full path
		surl, err := url.JoinPath(ss.baseURL, rurl.Alias)
		if err != nil {
			return nil, status.Error(codes.Internal, "internal error")
			continue
		}

		batchResp[i] = &pb.BatchURLItem{
			CorrelationId: breq.CorrelationId,
			Url:           surl,
		}
	}

	return &pb.BatchURLResponse{ShortUrls: batchResp}, nil
}

// GetUserURL returns user's urls
func (ss *ShortenerServer) GetUserURL(ctx context.Context, req *pb.GetUserURLRequest) (*pb.GetUserURLResponse, error) {
	// try to get user id from context
	userID, err := ss.getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userid is missing")
	}

	uurls, err := ss.storage.GetUserURLsCtx(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	// get number of user's url
	numUserUrls := len(uurls)
	// user url is not exist
	if numUserUrls == 0 {
		return nil, status.Error(codes.NotFound, "can't get user urls")
	}
	// output user urls
	userURLResp := make([]*pb.UserURLItem, 0)
	// prepare user urls
	for i, uurl := range uurls {
		urlPath, err := url.JoinPath(ss.baseURL, uurl.Alias)
		if err != nil {
			logger.Log.Error("join url path", zap.Error(err))
			continue
		}
		// forming items
		userURLResp[i] = &pb.UserURLItem{
			ShortUrl:    urlPath,
			OriginalUrl: uurl.URL,
		}
	}

	return &pb.GetUserURLResponse{Urls: userURLResp}, nil
}

// DeleteUserURL deletes user's urls
func (ss *ShortenerServer) DeleteUserURL(ctx context.Context, req *pb.DeleteUserURLRequest) (*pb.DeleteUserURLResponse, error) {
	// try to get user id from context
	userID, err := ss.getUserIDFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "userid is missing")
	}

	if len(req.ShortUrls) == 0 {
		return nil, status.Error(codes.InvalidArgument, "nothing to delete")
	}
	// pass user aliases to delete worker
	ss.urlsToDeleteCh <- models.UserDeleteTask{
		UID:     userID,
		Aliases: req.ShortUrls,
	}

	return &pb.DeleteUserURLResponse{}, nil
}

// Stats returns the number of shortened urls and users in the service
func (ss *ShortenerServer) Stats(ctx context.Context, req *pb.StatsRequest) (*pb.StatsResponse, error) {
	if ss.trustedNet == nil {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}
	// try to get client ip from context metadata
	clientIP, err := ss.getClientIPFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}
	// check trusted subnet includes client ip
	if !ss.trustedNet.Contains(clientIP) {
		return nil, status.Error(codes.PermissionDenied, "permission denied")
	}
	// get the number of shortened urls
	urls, err := ss.storage.GetURLCountCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}
	// get user count
	users, err := ss.storage.GetUserCountCtx(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &pb.StatsResponse{
		Urls:  int32(urls),
		Users: int32(users),
	}, nil
}

// Ping verifies a connection to the database
func (ss *ShortenerServer) Ping(ctx context.Context, req *pb.PingRequest) (*pb.PingResponse, error) {
	if dbStorage, ok := ss.storage.(*storage.DBStorage); ok {
		if err := dbStorage.Ping(ctx); err != nil {
			return &pb.PingResponse{Ok: false}, status.Error(codes.Unavailable, "can not connect to database")
		}
	}
	return &pb.PingResponse{Ok: true}, nil
}
