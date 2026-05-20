package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"api-gateway/internal/clients"
	"api-gateway/internal/config"
	"api-gateway/internal/handler"
	"api-gateway/internal/middleware"
	"api-gateway/internal/obs"
)

func Run() error {
	cfg := config.Load()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownTracer, err := obs.InitTracer(ctx, "api-gateway", cfg.OTLPEndpoint)
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

	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})
	defer rdb.Close()

	cs, err := clients.Dial(ctx, cfg.TrainServiceAddr, cfg.UserServiceAddr, cfg.BookingServiceAddr, cfg.NotificationServiceAddr)
	if err != nil {
		return fmt.Errorf("dial services: %w", err)
	}
	defer cs.Close()

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"GET", "POST", "PATCH", "DELETE", "PUT", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}))
	r.Use(middleware.RateLimit(rdb, cfg.RateLimitRPM, time.Minute))

	r.GET("/healthz", func(c *gin.Context) { c.JSON(200, gin.H{"ok": true}) })

	api := r.Group("/v1")
	publicAPI := api.Group("")
	authedAPI := api.Group("")
	authedAPI.Use(middleware.JWT(cfg.JWTSecret))

	userH := handler.NewUserHandler(cs.User)
	userH.RegisterPublic(publicAPI)
	userH.RegisterAuthenticated(authedAPI)

	trainH := handler.NewTrainHandler(cs.Train)
	trainH.RegisterPublic(publicAPI)
	trainH.RegisterAuthenticated(authedAPI)

	bookingH := handler.NewBookingHandler(cs.Booking)
	bookingH.Register(authedAPI)

	notificationH := handler.NewNotificationHandler(cs.Notification)
	notificationH.Register(authedAPI)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           otelhttp.NewHandler(r, "api-gateway"),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
	}

	go func() {
		log.Printf("api-gateway listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("http listen: %v", err)
			stop()
		}
	}()

	<-ctx.Done()
	log.Printf("shutting down gateway...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func MustRun() {
	if err := Run(); err != nil {
		_, _ = os.Stderr.WriteString("fatal: " + err.Error() + "\n")
		os.Exit(1)
	}
}
