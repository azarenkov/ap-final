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

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	gcreds "google.golang.org/grpc/credentials/insecure"

	bookingv1 "github.com/azarenkov/ap2-final-gen/booking/v1"
	userv1 "github.com/azarenkov/ap2-final-gen/user/v1"
	"user-service/internal/config"
	"user-service/internal/jwt"
	"user-service/internal/obs"
	"user-service/internal/repository/postgres"
	transportgrpc "user-service/internal/transport/grpc"
	"user-service/internal/usecase"
)

func Run() error {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTracer, err := obs.InitTracer(ctx, "user-service", cfg.OTLPEndpoint)
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

	bookingAddr := os.Getenv("BOOKING_SERVICE_ADDR")
	if bookingAddr == "" {
		bookingAddr = "booking-service:50053"
	}
	bConn, _ := grpc.NewClient(bookingAddr,
		grpc.WithTransportCredentials(gcreds.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	defer func() {
		if bConn != nil {
			_ = bConn.Close()
		}
	}()
	var bookingClient bookingv1.BookingServiceClient
	if bConn != nil {
		bookingClient = bookingv1.NewBookingServiceClient(bConn)
	}

	repo := postgres.NewUserRepo(db)
	issuer := jwt.NewIssuer(cfg.JWTSecret, cfg.JWTExpiry)
	uc := usecase.New(repo, issuer)
	srv := transportgrpc.NewServer(uc, bookingClient)

	lis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	grpcSrv := grpc.NewServer(grpc.StatsHandler(otelgrpc.NewServerHandler()))
	userv1.RegisterUserServiceServer(grpcSrv, srv)

	go func() {
		log.Printf("user-service grpc listening on %s", cfg.GRPCAddr)
		if err := grpcSrv.Serve(lis); err != nil {
			log.Printf("grpc serve: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Printf("user-service shutting down")
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
