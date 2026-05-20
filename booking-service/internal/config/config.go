package config

import (
	"os"
	"time"
)

type Config struct {
	GRPCAddr         string
	MetricsAddr      string
	DatabaseURL      string
	TrainServiceAddr string
	NATSURL          string
	OTLPEndpoint     string
	ShutdownGrace    time.Duration
}

func Load() Config {
	return Config{
		GRPCAddr:         env("BOOKING_GRPC_ADDR", ":50053"),
		MetricsAddr:      env("BOOKING_METRICS_ADDR", ":9093"),
		DatabaseURL:      env("BOOKING_DATABASE_URL", "postgres://postgres:postgres@localhost:5434/booking_db?sslmode=disable"),
		TrainServiceAddr: env("TRAIN_SERVICE_ADDR", "train-service:50051"),
		NATSURL:          env("NATS_URL", "nats://localhost:4222"),
		OTLPEndpoint:     env("OTLP_ENDPOINT", ""),
		ShutdownGrace:    10 * time.Second,
	}
}

func env(k, fb string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return fb
}
