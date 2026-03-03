package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"go.uber.org/zap"

	"github.com/jambotails/shipping-service/config"
	"github.com/jambotails/shipping-service/internal/cache"
	"github.com/jambotails/shipping-service/internal/handlers"
	"github.com/jambotails/shipping-service/internal/middleware"
	"github.com/jambotails/shipping-service/internal/repositories"
	"github.com/jambotails/shipping-service/internal/services"
)

func main() {
	// ── Load config ──────────────────────────────────────────
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// ── Logger ───────────────────────────────────────────────
	var logger *zap.Logger
	if cfg.Server.GinMode == "release" {
		logger, _ = zap.NewProduction()
	} else {
		logger, _ = zap.NewDevelopment()
	}
	defer logger.Sync()

	// ── Database ─────────────────────────────────────────────
	db, err := sql.Open("postgres", cfg.Database.DSN())
	if err != nil {
		logger.Fatal("failed to open database", zap.Error(err))
	}
	defer db.Close()

	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetimeMin)

	if err := db.Ping(); err != nil {
		logger.Fatal("failed to ping database", zap.Error(err))
	}
	logger.Info("database connected")

	// ── Redis ────────────────────────────────────────────────
	redisClient, err := cache.NewRedisClient(
		cfg.Redis.URL, cfg.Redis.Addr(), cfg.Redis.Password, cfg.Redis.DB, logger,
	)
	if err != nil {
		logger.Warn("redis init error (cache disabled)", zap.Error(err))
	}
	defer redisClient.Close()

	// ── Repositories ─────────────────────────────────────────
	warehouseRepo := repositories.NewWarehouseRepository(db)
	customerRepo := repositories.NewCustomerRepository(db)
	sellerRepo := repositories.NewSellerRepository(db)
	productRepo := repositories.NewProductRepository(db)
	transportRateRepo := repositories.NewTransportRateRepository(db)
	speedConfigRepo := repositories.NewDeliverySpeedConfigRepository(db)

	// ── Services ─────────────────────────────────────────────
	warehouseSvc := services.NewWarehouseService(warehouseRepo, sellerRepo, productRepo)
	shippingSvc := services.NewShippingService(
		warehouseRepo, customerRepo, productRepo,
		transportRateRepo, speedConfigRepo, warehouseSvc,
	)

	// ── Handlers ─────────────────────────────────────────────
	warehouseHandler := handlers.NewWarehouseHandler(warehouseSvc)
	shippingHandler := handlers.NewShippingHandler(shippingSvc)
	healthHandler := handlers.NewHealthHandler(db, redisClient)

	// ── Gin engine ───────────────────────────────────────────
	gin.SetMode(cfg.Server.GinMode)
	r := gin.New()

	// CORS — allow all origins for public API.
	// For production, restrict via CORS_ALLOWED_ORIGINS env (comma-separated);
	// if unset, defaults to wide-open (suitable for a demo/assignment deployment).
	var allowedOrigins []string
	if cfg.Server.CORSOrigins != "" {
		for _, o := range strings.Split(cfg.Server.CORSOrigins, ",") {
			if o = strings.TrimSpace(o); o != "" {
				allowedOrigins = append(allowedOrigins, o)
			}
		}
	}

	corsConf := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
	if len(allowedOrigins) > 0 {
		corsConf.AllowOrigins = allowedOrigins
	} else {
		// No restriction — allow every origin (Vercel previews, localhost, etc.)
		corsConf.AllowAllOrigins = true
	}
	r.Use(cors.New(corsConf))

	// Global middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recovery(logger))
	r.Use(middleware.RateLimiter(100, 100.0/60.0)) // 100 tokens, refill ~1.67/sec

	// ── Routes ───────────────────────────────────────────────
	r.GET("/health", healthHandler.Check)

	v1 := r.Group("/api/v1")
	{
		// Optionally protect routes with JWT:
		// v1.Use(middleware.Auth(cfg.JWT.Secret))

		v1.GET("/warehouse/nearest", warehouseHandler.FindNearest)
		v1.GET("/shipping-charge", shippingHandler.GetCharge)
		v1.POST("/shipping-charge/calculate", shippingHandler.CalculateFull)
	}

	// ── HTTP Server with graceful shutdown ───────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("server starting", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited")
}
