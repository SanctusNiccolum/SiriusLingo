package service

import (
	"context"
	"time"

	pb "github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/gen/go/proto"
	"github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/internal/config"
	"github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/internal/db"
	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
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

func (s *AuthService) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	s.logger.Debug("Logging in user", zap.String("username", req.Username))
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	user, err := s.db.UserQuery().GetByUsername(ctx, req.Username)
	if err != nil {
		s.logger.Error("Failed to fetch user", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to fetch user")
	}
	if user == nil {
		s.logger.Warn("User not found", zap.String("username", req.Username))
		return nil, status.Error(codes.NotFound, "user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.Warn("Invalid password", zap.String("username", req.Username))
		return nil, status.Error(codes.Unauthenticated, "invalid password")
	}

	role, err := s.db.RoleQuery().GetByID(ctx, user.RoleID)
	if err != nil {
		s.logger.Error("Failed to fetch role", zap.Error(err), zap.Int64("role_id", user.RoleID))
		return nil, status.Error(codes.Internal, "failed to fetch role")
	}
	if role == nil {
		s.logger.Warn("Role not found", zap.Int64("role_id", user.RoleID))
		return nil, status.Error(codes.NotFound, "role not found")
	}

	accessJTI := uuid.New().String()
	refreshJTI := uuid.New().String()

	accessToken, err := s.generateJWT(user.ID, "access", role.Name, s.config.ACCESS_TOKEN_EXPIRES_IN, []byte(user.AccessTokenSecret), accessJTI)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to generate access token")
	}
	refreshToken, err := s.generateJWT(user.ID, "refresh", role.Name, s.config.REFRESH_TOKEN_EXPIRES_IN, []byte(user.RefreshTokenSecret), refreshJTI)
	if err != nil {
		s.logger.Error("Failed to generate refresh token", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to generate refresh token")
	}

	user.AccessTokenJTI = &accessJTI
	user.RefreshTokenJTI = &refreshJTI
	_, err = s.db.UserQuery().UpdateLoginOrLogout(ctx, user, user.ID)
	_, err = s.db.UserQuery().UpdateAuthTime(ctx, user.ID)
	if err != nil {
		s.logger.Error("Failed to update token JTI", zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to update token JTI")
	}

	s.logger.Info("User logged in successfully", zap.Int64("user_id", user.ID), zap.String("username", req.Username))
	return &pb.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) generateJWT(userID int64, tokenType string, roleName string, expiresIn time.Duration, secretKey []byte, jti string) (string, error) {
	claims := jwt.MapClaims{
		"sub":  userID,
		"type": tokenType,
		"role": roleName,
		"exp":  time.Now().Add(expiresIn).Unix(),
		"iat":  time.Now().Unix(),
		"jti":  jti,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secretKey)
}
