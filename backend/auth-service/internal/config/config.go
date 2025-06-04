package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	DBHost                   string
	DBPort                   string
	DBUser                   string
	DBPassword               string
	DBName                   string
	GRPCAddr                 string
	ACCESS_TOKEN_EXPIRES_IN  time.Duration
	REFRESH_TOKEN_EXPIRES_IN time.Duration
}

func LoadConfig() (*AppConfig, error) {
	err := godotenv.Load("./config/.env")
	if err != nil {
		fmt.Printf("Failed to load .env file: %v\n", err)
	}

	cfg := AppConfig{
		DBHost:     os.Getenv("DB_HOST"),
		DBPort:     os.Getenv("DB_PORT"),
		DBUser:     os.Getenv("DB_USER"),
		DBPassword: os.Getenv("DB_PASSWORD"),
		DBName:     os.Getenv("DB_NAME"),
		GRPCAddr:   os.Getenv("GRPC_ADDR"),
	}

	accessTokenExpiresIn := os.Getenv("ACCESS_TOKEN_EXPIRES_IN")
	if accessTokenExpiresIn == "" {
		return nil, fmt.Errorf("ACCESS_TOKEN_EXPIRES_IN is empty or not set in .env")
	}
	accessDuration, err := time.ParseDuration(accessTokenExpiresIn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ACCESS_TOKEN_EXPIRES_IN: %v", err)
	}
	cfg.ACCESS_TOKEN_EXPIRES_IN = accessDuration

	refreshTokenExpiresIn := os.Getenv("REFRESH_TOKEN_EXPIRES_IN")
	if refreshTokenExpiresIn == "" {
		return nil, fmt.Errorf("REFRESH_TOKEN_EXPIRES_IN is empty or not set in .env")
	}
	refreshDuration, err := time.ParseDuration(refreshTokenExpiresIn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse REFRESH_TOKEN_EXPIRES_IN: %v", err)
	}
	cfg.REFRESH_TOKEN_EXPIRES_IN = refreshDuration

	return &cfg, nil
}
