package config

import (
	"os"
	"time"
)

type Config struct {
	GRPCAddr      string
	MetricsAddr   string
	DatabaseURL   string
	RedisAddr     string
	NATSURL       string
	OTLPEndpoint  string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	ShutdownGrace time.Duration
}

func Load() Config {
	return Config{
		GRPCAddr:      env("TRAIN_GRPC_ADDR", ":50051"),
		MetricsAddr:   env("TRAIN_METRICS_ADDR", ":9091"),
		DatabaseURL:   env("TRAIN_DATABASE_URL", "postgres://postgres:postgres@localhost:5432/train_db?sslmode=disable"),
		RedisAddr:     env("REDIS_ADDR", "localhost:6379"),
		NATSURL:       env("NATS_URL", "nats://localhost:4222"),
		OTLPEndpoint:  env("OTLP_ENDPOINT", ""),
		ReadTimeout:   5 * time.Second,
		WriteTimeout:  10 * time.Second,
		ShutdownGrace: 10 * time.Second,
	}
}

func env(k, fb string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return fb
}
