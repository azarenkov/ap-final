package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	gcreds "google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"

	notificationv1 "github.com/azarenkov/ap2-final-gen/notification/v1"
	userv1 "github.com/azarenkov/ap2-final-gen/user/v1"
	"notification-service/internal/config"
	"notification-service/internal/events"
	"notification-service/internal/obs"
	"notification-service/internal/repository/postgres"
	"notification-service/internal/sender"
	transportgrpc "notification-service/internal/transport/grpc"
	"notification-service/internal/usecase"
)

func Run() error {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTracer, err := obs.InitTracer(ctx, "notification-service", cfg.OTLPEndpoint)
	if err != nil {
		log.Printf("tracer init: %v (continuing without traces)", err)
		shutdownTracer = func(context.Context) error { return nil }
	}
	defer func() {
		sc, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = shutdownTracer(sc)
	}()
	obs.ServeMetrics(ctx, cfg.MetricsAddr)

	db, err := openDB(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer db.Close()

	repo := postgres.New(db)
	var emailer sender.EmailSender
	if cfg.SMTPHost != "" {
		emailer = sender.NewSMTPSender(cfg.SMTPHost, cfg.SMTPFrom, cfg.SMTPUsername, cfg.SMTPPassword, cfg.SMTPUseTLS)
		log.Printf("notification-service emailing via SMTP %s tls=%v as %s", cfg.SMTPHost, cfg.SMTPUseTLS, cfg.SMTPFrom)
	} else {
		emailer = sender.NewLogSender(cfg.SMTPFrom)
		log.Printf("notification-service emailing via LOG (no MAILER_HOST set)")
	}
	uc := usecase.New(repo, emailer)

	nc, err := nats.Connect(cfg.NATSURL, nats.Timeout(5*time.Second), nats.ReconnectWait(2*time.Second), nats.MaxReconnects(-1))
	if err != nil {
		log.Printf("nats connect failed: %v", err)
	} else {
		defer nc.Drain()
	}

	var resolver events.EmailResolver
	if cfg.UserServiceAddr != "" {
		uConn, err := grpc.NewClient(cfg.UserServiceAddr,
			grpc.WithTransportCredentials(gcreds.NewCredentials()),
			grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		)
		if err != nil {
			log.Printf("user-service dial failed: %v", err)
		} else {
			defer uConn.Close()
			resolver = events.NewUserServiceResolver(userv1.NewUserServiceClient(uConn))
			log.Printf("notification-service email-resolver wired to user-service %s", cfg.UserServiceAddr)
		}
	}

	sub := events.NewSubscriber(uc, resolver)
	if err := sub.Start(nc); err != nil {
		log.Printf("event subscribe failed: %v", err)
	}
	defer sub.Stop()

	srv := transportgrpc.NewServer(uc)
	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	grpcSrv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	notificationv1.RegisterNotificationServiceServer(grpcSrv, srv)
	reflection.Register(grpcSrv)

	go func() {
		log.Printf("notification-service grpc listening on %s", cfg.GRPCAddr)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Printf("grpc serve: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Printf("notification-service shutting down")
	done := make(chan struct{})
	go func() {
		grpcSrv.GracefulStop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(cfg.ShutdownGrace):
		grpcSrv.Stop()
	}
	return nil
}

func openDB(ctx context.Context, url string) (*sql.DB, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func MustRun() {
	if err := Run(); err != nil {
		_, _ = os.Stderr.WriteString("fatal: " + err.Error() + "\n")
		os.Exit(1)
	}
}
