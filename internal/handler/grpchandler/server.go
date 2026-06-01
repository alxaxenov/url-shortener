package grpchandler

import (
	"context"
	"errors"
	"fmt"
	"net"

	"github.com/alxaxenov/url-shortener/tree/v2/internal/handler/grpchandler/pb"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/logger"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/model"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/service"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/utils"
	"github.com/alxaxenov/url-shortener/tree/v2/internal/worker/audit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

type IShortenerService interface {
	AddURL(context.Context, string, int) (string, error)
	GetURL(context.Context, string) (string, error)
	UserURLs(context.Context, int) ([]model.UserURLs, error)
}

type AuditPublisher interface {
	Publish(action audit.ActionType, userID int, URL string)
}

type ITokenManager interface {
	GetUserID(tokenString string) (int, error)
}

type ShortenerServer struct {
	pb.UnimplementedShortenerServiceServer

	service IShortenerService
	audit   AuditPublisher
}

func NewGRPCServer(addr string, srv IShortenerService, audit AuditPublisher, tokenManager ITokenManager) (*grpc.Server, net.Listener, error) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, nil, fmt.Errorf("NewGRPCServer Listen error %w", err)
	}
	userInterceptor := UserIDInterceptor(tokenManager)
	s := grpc.NewServer(grpc.UnaryInterceptor(userInterceptor))
	pb.RegisterShortenerServiceServer(s, &ShortenerServer{service: srv, audit: audit})
	return s, lis, nil
}

func (s *ShortenerServer) ShortenURL(ctx context.Context, in *pb.URLShortenRequest) (*pb.URLShortenResponse, error) {
	origin := in.GetUrl()

	userID, err := utils.GetUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	var respErr error
	short, err := s.service.AddURL(ctx, origin, userID)
	if err != nil {
		var badURL *service.BadURL
		var alreadyExists *service.AlreadyExists
		if errors.As(err, &alreadyExists) {
			respErr = status.Errorf(codes.AlreadyExists, `URL %s already exist`, origin)
		} else if errors.As(err, &badURL) {
			logger.Logger.Info("gRPC AddURL service.AddURL badURL", "error", err)
			return nil, status.Errorf(codes.InvalidArgument, `incorrect URL %s`, origin)
		} else {
			logger.Logger.Error("gRPC AddURL service.AddURL", "error", err)
			return nil, status.Error(codes.Internal, `Internal error`)
		}
	}
	if respErr == nil && s.audit != nil {
		go s.audit.Publish(audit.Shorten, userID, origin)
	}

	return pb.URLShortenResponse_builder{Result: proto.String(short)}.Build(), respErr
}

func (s *ShortenerServer) ExpandURL(ctx context.Context, in *pb.URLExpandRequest) (*pb.URLExpandResponse, error) {
	origin, err := s.service.GetURL(ctx, in.GetId())
	if err != nil {
		var er error
		if errors.Is(err, service.ErrURLDeleted) {
			er = status.Errorf(codes.NotFound, `URL %s is not exist or is deleted`, in.GetId())
		} else {
			logger.Logger.Error("gRPC GetValue service.GetURL", "error", err)
			er = status.Errorf(codes.Internal, `Internal error`)
		}
		return nil, er
	}
	if s.audit != nil {
		userID, _ := utils.GetUserID(ctx)
		go s.audit.Publish(audit.Follow, userID, origin)
	}

	return pb.URLExpandResponse_builder{Result: proto.String(origin)}.Build(), nil
}

func (s *ShortenerServer) ListUserURLs(ctx context.Context, _ *emptypb.Empty) (*pb.UserURLsResponse, error) {
	userID, err := utils.GetUserID(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "userId context find error")
	}

	data, err := s.service.UserURLs(ctx, userID)
	if err != nil {
		logger.Logger.Error("UserURLs service.UserUrls", "error", err)
		return nil, status.Errorf(codes.Internal, `Internal error`)
	}
	repeated := make([]*pb.URLData, 0, len(data))
	for _, u := range data {
		repeated = append(repeated, pb.URLData_builder{ShortUrl: &u.Short, OriginalUrl: &u.Origin}.Build())
	}
	return pb.UserURLsResponse_builder{Url: repeated}.Build(), nil
}

func UserIDInterceptor(tokenManager ITokenManager) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok || len(md.Get("authorization")) == 0 {
			return nil, status.Error(codes.Unauthenticated, `missing token`)
		}
		values := md.Get("authorization")
		userID, err := tokenManager.GetUserID(values[0])
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, `invalid token`)
		}
		userCtx := utils.SetUserID(ctx, userID)
		return handler(userCtx, req)
	}
}
