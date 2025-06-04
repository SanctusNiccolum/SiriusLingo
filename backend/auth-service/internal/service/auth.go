package service

import (
	"context"
	"time"

	pb "github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/gen/go/proto"
	"github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/internal/config"
	"github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/internal/db"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const DefaultRoleName = "user"

type AuthService struct {
	pb.UnimplementedAuthServiceServer
	db     db.Implementation
	logger *zap.Logger
	config config.AppConfig
}

func NewAuthService(db db.Implementation, logger *zap.Logger, cfg config.AppConfig) *AuthService {
	return &AuthService{
		db:     db,
		logger: logger,
		config: cfg,
	}
}

func (s *AuthService) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	s.logger.Debug("Registering new user", zap.String("username", req.Username))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	exists, err := s.db.UserQuery().ExistsByUsernameOrEmail(ctx, req.Username, req.Email)
	if err != nil {
		s.logger.Error("Failed to check uniqueness", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to check uniqueness")
	}
	if exists {
		s.logger.Warn("Username or email already exists",
			zap.String("username", req.Username),
			zap.String("email", req.Email))
		return nil, status.Error(codes.AlreadyExists, "username or email already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	accessTokenSecret, err := db.GenerateSecretKey()
	if err != nil {
		s.logger.Error("Failed to generate access token secret", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to generate access token secret")
	}
	refreshTokenSecret, err := db.GenerateSecretKey()
	if err != nil {
		s.logger.Error("Failed to generate refresh token secret", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to generate refresh token secret")
	}
	DefaultRoleID, err := s.db.RoleQuery().GetIDByName(ctx, DefaultRoleName)
	if err != nil {
		s.logger.Error("Failed to get default role ID", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to get default role ID")
	}
	newUser := &db.User{
		Username:           req.Username,
		Password:           string(hashedPassword),
		Email:              req.Email,
		RoleID:             DefaultRoleID,
		AccessTokenSecret:  accessTokenSecret,
		RefreshTokenSecret: refreshTokenSecret,
	}

	_, err = s.db.UserQuery().Insert(ctx, newUser)
	if err != nil {
		s.logger.Error("Failed to insert user", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to insert user")
	}

	s.logger.Info("User registered successfully", zap.String("username", req.Username))
	return &pb.RegisterResponse{}, nil
}
