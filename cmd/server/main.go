package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/geo-service/internal/api"
	"github.com/geo-service/internal/config"
	"github.com/geo-service/internal/geoip"
	"github.com/geo-service/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	httpPort  = flag.String("http-port", "8080", "HTTP server port")
	grpcPort  = flag.String("grpc-port", "9090", "gRPC server port")
	dbPath    = flag.String("db-path", "./data/GeoLite2-Country.mmdb", "Path to MaxMind database")
	logLevel  = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	enableTLS = flag.Bool("enable-tls", false, "Enable TLS for gRPC")
)

func main() {
	flag.Parse()

	// Setup logger
	logger := setupLogger(*logLevel)

	// Load configuration
	cfg := &config.Config{
		HTTPPort:    *httpPort,
		GRPCPort:    *grpcPort,
		DatabasePath: *dbPath,
		EnableTLS:   *enableTLS,
	}

	// Initialize GeoIP service
	geoService, err := geoip.NewService(cfg.DatabasePath, logger)
	if err != nil {
		logger.Fatalf("Failed to initialize GeoIP service: %v", err)
	}
	defer geoService.Close()

	// Create API handlers
	apiHandler := api.NewHandler(geoService, logger)

	// Setup HTTP server
	httpServer := setupHTTPServer(cfg.HTTPPort, apiHandler, logger)

	// Setup gRPC server
	grpcServer := setupGRPCServer(cfg.GRPCPort, apiHandler, logger)

	// Start servers
	go func() {
		logger.Infof("Starting HTTP server on port %s", cfg.HTTPPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("HTTP server failed: %v", err)
		}
	}()

	go func() {
		logger.Infof("Starting gRPC server on port %s", cfg.GRPCPort)
		lis, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			logger.Fatalf("Failed to listen: %v", err)
		}
		if err := grpcServer.Serve(lis); err != nil {
			logger.Fatalf("gRPC server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down servers...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		logger.Errorf("HTTP server shutdown failed: %v", err)
	}

	grpcServer.GracefulStop()
	logger.Info("Servers shut down successfully")
}

func setupLogger(level string) *logrus.Logger {
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logger.Warnf("Invalid log level %s, defaulting to info", level)
		logLevel = logrus.InfoLevel
	}
	logger.SetLevel(logLevel)

	return logger
}

func setupHTTPServer(port string, handler *api.Handler, logger *logrus.Logger) *http.Server {
	r := mux.NewRouter()

	// Register middleware
	r.Use(middleware.Logging(logger))
	r.Use(middleware.Recovery(logger))

	// Register routes
	handler.RegisterHTTPRoutes(r)

	return &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
}

func setupGRPCServer(port string, handler *api.Handler, logger *logrus.Logger) *grpc.Server {
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(middleware.GRPCLogging(logger)),
	}

	server := grpc.NewServer(opts...)
	handler.RegisterGRPCServer(server)

	return server
}