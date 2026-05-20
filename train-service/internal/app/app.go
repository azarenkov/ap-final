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
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"

	trainv1 "github.com/azarenkov/ap2-final-gen/train/v1"
	"train-service/internal/cache"
	"train-service/internal/config"
	"train-service/internal/events"
	"train-service/internal/obs"
	"train-service/internal/repository/postgres"
	transportgrpc "train-service/internal/transport/grpc"
	"train-service/internal/usecase"
)

func Run() error {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTracer, err := obs.InitTracer(ctx, "train-service", cfg.OTLPEndpoint)
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

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer rdb.Close()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("redis ping failed (continuing without cache): %v", err)
	}

	nc, err := nats.Connect(cfg.NATSURL, nats.Timeout(5*time.Second))
	if err != nil {
		log.Printf("nats connect failed (events will be dropped): %v", err)
	} else {
		defer nc.Drain()
	}

	trainRepo := postgres.NewTrainRepo(db)
	routeRepo := postgres.NewRouteRepo(db)
	trainCache := cache.NewTrainCache(rdb)
	publisher := events.NewNATSPublisher(nc)

	uc := usecase.New(trainRepo, routeRepo, trainCache, publisher)
	server := transportgrpc.NewTrainServer(uc)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", cfg.GRPCAddr, err)
	}

	grpcSrv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	trainv1.RegisterTrainServiceServer(grpcSrv, server)

	go func() {
		log.Printf("train-service grpc listening on %s", cfg.GRPCAddr)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Printf("grpc serve: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Printf("shutting down...")
	doneCh := make(chan struct{})
	go func() {
		grpcSrv.GracefulStop()
		close(doneCh)
	}()
	select {
	case <-doneCh:
	case <-time.After(cfg.ShutdownGrace):
		log.Printf("graceful stop timeout — forcing")
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
