package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/internal/config"
	"github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/internal/deps"
	"github.com/SanctusNiccolum/SiriusLingo/backend/auth-service/internal/logger"
	"go.uber.org/zap"
)

func main() {
	log := logger.NewLogger()
	defer log.Sync()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Error loading config", zap.Error(err))
	}
	if cfg.DBHost == "" || cfg.DBPort == "" || cfg.DBUser == "" || cfg.DBPassword == "" || cfg.DBName == "" || cfg.GRPCAddr == "" {
		log.Fatal("Missing required configuration values in .env file")
		return
	}

	depends, err := deps.ProvideDependencies(*cfg)
	if err != nil {
		log.Fatal("Failed to initialize dependencies", zap.Error(err))
	}
	defer depends.Cleanup()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Server started successfully", zap.String("address", cfg.GRPCAddr))
	<-sigChan

	log.Info("Shutting down server...")
}
