package config

import (
	"os"
	"strconv"
)

type Config struct {
	HTTPAddr                string
	MetricsAddr             string
	TrainServiceAddr        string
	UserServiceAddr         string
	BookingServiceAddr      string
	NotificationServiceAddr string
	JWTSecret               string
	RateLimitRPM            int
	RedisAddr               string
	OTLPEndpoint            string
}

func Load() Config {
	return Config{
		HTTPAddr:                env("GATEWAY_HTTP_ADDR", ":8080"),
		MetricsAddr:             env("GATEWAY_METRICS_ADDR", ":9095"),
		TrainServiceAddr:        env("TRAIN_SERVICE_ADDR", "localhost:50051"),
		UserServiceAddr:         env("USER_SERVICE_ADDR", "localhost:50052"),
		BookingServiceAddr:      env("BOOKING_SERVICE_ADDR", "localhost:50053"),
		NotificationServiceAddr: env("NOTIFICATION_SERVICE_ADDR", "localhost:50054"),
		JWTSecret:               env("JWT_SECRET", "dev-secret-change-me"),
		RateLimitRPM:            envInt("RATE_LIMIT_RPM", 600),
		RedisAddr:               env("REDIS_ADDR", "localhost:6379"),
		OTLPEndpoint:            env("OTLP_ENDPOINT", ""),
	}
}

func env(k, fb string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return fb
}

func envInt(k string, fb int) int {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fb
}
