package config

import (
	"errors"
	"os"
)

type Config struct {
	Port string
	DatabaseURL string
	JWTSecret   string
}

func Load() (Config, error)  {
	cfg := Config{
		Port: getEnv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret : os.Getenv("JWT_SECRET"),
	}

	

	if cfg.DatabaseURL == ""{
		return Config{}, errors.New("DATABASE_URL is required")
	}

	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET is required and must be at least 32 characters")
	}

	return cfg, nil

}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}