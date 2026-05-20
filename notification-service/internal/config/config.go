package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	GRPCAddr        string
	MetricsAddr     string
	DatabaseURL     string
	NATSURL         string
	UserServiceAddr string
	SMTPHost        string
	SMTPFrom        string
	SMTPUsername    string
	SMTPPassword    string
	SMTPUseTLS      bool
	OTLPEndpoint    string
	ShutdownGrace   time.Duration
}

func Load() Config {
	host := env("MAILER_HOST", env("SMTP_HOST", ""))
	port := env("MAILER_PORT", "")
	if host != "" && port != "" && !strings.Contains(host, ":") {
		host = host + ":" + port
	}
	useTLS := strings.EqualFold(env("MAILER_TLS", ""), "true") || strings.HasSuffix(host, ":465")

	return Config{
		GRPCAddr:        env("NOTIFICATION_GRPC_ADDR", ":50054"),
		MetricsAddr:     env("NOTIFICATION_METRICS_ADDR", ":9094"),
		DatabaseURL:     env("NOTIFICATION_DATABASE_URL", "postgres://postgres:postgres@localhost:5435/notification_db?sslmode=disable"),
		NATSURL:         env("NATS_URL", "nats://localhost:4222"),
		UserServiceAddr: env("USER_SERVICE_ADDR", "user-service:50052"),
		SMTPHost:        host,
		SMTPFrom:        env("MAILER_FROM", env("SMTP_FROM", "no-reply@trains.test")),
		SMTPUsername:    env("MAILER_USERNAME", env("SMTP_USERNAME", "")),
		SMTPPassword:    env("MAILER_PASSWORD", env("SMTP_PASSWORD", "")),
		SMTPUseTLS:      useTLS,
		OTLPEndpoint:    env("OTLP_ENDPOINT", ""),
		ShutdownGrace:   10 * time.Second,
	}
}

func env(k, fb string) string {
	if v, ok := os.LookupEnv(k); ok && v != "" {
		return v
	}
	return fb
}
