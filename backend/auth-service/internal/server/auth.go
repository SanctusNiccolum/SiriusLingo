package server

import (
	"context"
	"net"

	pb "github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/gen/go/proto"
	"github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/internal/service"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type AuthServer struct {
	pb.UnimplementedAuthServiceServer
	grpcServer *grpc.Server
	errChan    chan error
	logger     *zap.Logger
	service    *service.AuthService
}

func NewAuthServer(svc *service.AuthService, logger *zap.Logger, addr string) (*AuthServer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()
	s := &AuthServer{
		grpcServer: grpcServer,
		errChan:    make(chan error, 1),
		logger:     logger,
		service:    svc,
	}
	reflection.Register(grpcServer)
	pb.RegisterAuthServiceServer(grpcServer, s)

	go func() {
		s.errChan <- grpcServer.Serve(listener)
	}()

	return s, nil
}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	return s.service.Register(ctx, req)
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	return s.service.Login(ctx, req)
}

func (s *AuthServer) ValidateToken(ctx context.Context, req *pb.ValidateTokenRequest) (*pb.ValidateTokenResponse, error) {
	s.logger.Debug("Validating token", zap.String("token_type", req.TokenType))
	_, err := s.service.ValidateToken(ctx, req.Token, req.TokenType)
	if err != nil {
		s.logger.Error("Token validation failed", zap.Error(err))
		return nil, err
	}
	s.logger.Info("Token validated successfully")
	return &pb.ValidateTokenResponse{}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	s.logger.Debug("Logging out user", zap.Int64("user_id", req.UserId))
	err := s.service.Logout(ctx, req.UserId)
	if err != nil {
		s.logger.Error("Logout failed", zap.Error(err))
		return nil, err
	}
	s.logger.Info("Logout successful", zap.Int64("user_id", req.UserId))
	return &pb.LogoutResponse{}, nil
}

func (s *AuthServer) ErrChan() chan error {
	return s.errChan
}

func (s *AuthServer) Stop() {
	s.grpcServer.GracefulStop()
}
