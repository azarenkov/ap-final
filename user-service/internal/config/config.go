package config

import (
	"os"
	"time"
)

type Config struct {
	GRPCAddr      string
	MetricsAddr   string
	DatabaseURL   string
	JWTSecret     string
	JWTExpiry     time.Duration
	OTLPEndpoint  string
	ShutdownGrace time.Duration
}

func Load() Config {
	return Config{
		GRPCAddr:      env("USER_GRPC_ADDR", ":50052"),
		MetricsAddr:   env("USER_METRICS_ADDR", ":9092"),
		DatabaseURL:   env("USER_DATABASE_URL", "postgres://postgres:postgres@localhost:5433/user_db?sslmode=disable"),
		JWTSecret:     env("JWT_SECRET", "dev-secret-change-me"),
		JWTExpiry:     24 * time.Hour,
		OTLPEndpoint:  env("OTLP_ENDPOINT", ""),
		ShutdownGrace: 10 * time.Second,
	}
}

func env(k, fb string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return fb
}
