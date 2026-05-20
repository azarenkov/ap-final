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

	"booking-service/internal/config"
	"booking-service/internal/events"
	"booking-service/internal/obs"
	"booking-service/internal/repository/postgres"
	transportgrpc "booking-service/internal/transport/grpc"
	"booking-service/internal/usecase"
	bookingv1 "github.com/azarenkov/ap2-final-gen/booking/v1"
	trainv1 "github.com/azarenkov/ap2-final-gen/train/v1"
)

func Run() error {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTracer, err := obs.InitTracer(ctx, "booking-service", cfg.OTLPEndpoint)
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

	tConn, err := grpc.NewClient(cfg.TrainServiceAddr,
		grpc.WithTransportCredentials(gcreds.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return fmt.Errorf("dial train: %w", err)
	}
	defer tConn.Close()
	trainClient := trainv1.NewTrainServiceClient(tConn)

	nc, err := nats.Connect(cfg.NATSURL, nats.Timeout(5*time.Second))
	if err != nil {
		log.Printf("nats connect failed: %v", err)
	} else {
		defer nc.Drain()
	}
	pub := events.NewPublisher(nc)

	repo := postgres.NewBookingRepo(db)
	uc := usecase.New(repo, trainClient, pub)
	srv := transportgrpc.NewServer(uc)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	grpcSrv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	bookingv1.RegisterBookingServiceServer(grpcSrv, srv)

	go func() {
		log.Printf("booking-service grpc listening on %s", cfg.GRPCAddr)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Printf("grpc serve: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Printf("booking-service shutting down")
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
