package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Env             string
	Port            string
	DatabaseURL     string
	JWTSecret       string
	LogLevel        string
	TrustedProxies  []string
	RequestTimeout  time.Duration
	ShutdownTimeout time.Duration
}

func Load() (Config, error)  {
	cfg := Config{
		Env:  getEnv("APP_ENV", "development"),
		Port: getEnv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret : os.Getenv("JWT_SECRET"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
		TrustedProxies: splitList(os.Getenv("TRUSTED_PROXIES")),
	}

	

	if cfg.DatabaseURL == ""{
		return Config{}, errors.New("DATABASE_URL is required")
	}

	if len(cfg.JWTSecret) < 32 {
		return Config{}, errors.New("JWT_SECRET is required and must be at least 32 characters")
	}

	var err error
	if cfg.RequestTimeout, err = getDuration("REQUEST_TIMEOUT", 10*time.Second); err != nil {
		return Config{}, err
	}
	if cfg.ShutdownTimeout, err = getDuration("SHUTDOWN_TIMEOUT", 20*time.Second); err != nil {
		return Config{}, err
	}

	return cfg, nil

}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}


func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	v := os.Getenv(key)
	if v == "" {
		return fallback, nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}
	return d, nil
}


func splitList(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}