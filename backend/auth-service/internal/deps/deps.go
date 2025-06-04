package deps

import (
	"github.com/Masterminds/squirrel"
	"github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/internal/config"
	"github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/internal/db"
	"github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/internal/logger"
	"github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/internal/server"
	"github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type Dependencies struct {
	Pool        *pgxpool.Pool
	DB          db.Implementation
	Logger      *zap.Logger
	AuthService *service.AuthService
	AuthServer  *server.AuthServer
}

func ProvideDependencies(cfg config.AppConfig) (*Dependencies, error) {
	log := logger.NewLogger()

	pool, err := db.InitDB(cfg, log)
	if err != nil {
		log.Fatal("Failed to init db", zap.Error(err))
		return nil, err
	}

	sq := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	deps := &Dependencies{
		Pool: pool,
		DB: db.NewImplementation(
			db.NewUserQuery(pool, sq, log),
			db.NewRoleQuery(pool, sq, log),
		),
		Logger: log,
	}

	deps.AuthService = service.NewAuthService(deps.DB, log, cfg)

	deps.AuthServer, err = server.NewAuthServer(deps.AuthService, log, cfg.GRPCAddr)
	if err != nil {
		log.Fatal("Failed to init auth server", zap.Error(err))
		pool.Close()
		return nil, err
	}

	go func() {
		if err := <-deps.AuthServer.ErrChan(); err != nil {
			log.Fatal("gRPC server failed", zap.Error(err))
		}
	}()

	log.Info("Dependencies initialized successfully")
	return deps, nil
}

func (d *Dependencies) Cleanup() {
	d.Logger.Info("Cleaning up dependencies")
	d.AuthServer.Stop()
	d.Logger.Sync()
	d.Pool.Close()
}
