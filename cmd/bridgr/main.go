package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/leofvo/bridgr/internal/config"
	"github.com/leofvo/bridgr/internal/domain"
	"github.com/leofvo/bridgr/internal/exporters"
	"github.com/leofvo/bridgr/internal/handlers"
	"github.com/leofvo/bridgr/internal/services"
	"github.com/leofvo/bridgr/internal/sources"
	"github.com/leofvo/bridgr/internal/store"
	"github.com/leofvo/bridgr/pkg/logger"
)

func main() {
	// Initialize logger
	if err := logger.Init("info"); err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.LoadConfig("/etc/bridgr/config.yaml")
	if err != nil {
		logger.Fatal("Failed to load configuration: %v", err)
	}

	// Validate configuration
	if err := config.ValidateConfig(cfg); err != nil {
		logger.Fatal("Invalid configuration: %v", err)
	}

	// Initialize Redis store
	redisStore, err := store.NewRedisStore(&cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to initialize Redis store: %v", err)
	}
	defer redisStore.Close()

	// Create factories
	sourceFactory := sources.NewFactory()
	exporterFactory := exporters.NewFactory()

	// Create sources and exporters
	var allSources []domain.Source
	var allExporters []domain.Exporter

	for _, group := range cfg.Groups {
		// Create sources
		for _, sourceCfg := range group.Sources {
			source, err := sourceFactory.CreateSource(&sourceCfg, group.Name)
			if err != nil {
				logger.Fatal("Failed to create source: error=%v group=%s", err, group.Name)
			}
			allSources = append(allSources, source)
		}

		// Create exporters
		for _, exporterCfg := range group.Exporters {
			exporter, err := exporterFactory.CreateExporter(&exporterCfg, group.Name)
			if err != nil {
				logger.Fatal("Failed to create exporter: error=%v group=%s", err, group.Name)
			}
			allExporters = append(allExporters, exporter)
		}
	}

	// Create services
	notificationService := services.NewNotificationService(redisStore)
	schedulerService := services.NewSchedulerService(notificationService, allSources, allExporters)

	// Create router
	router := mux.NewRouter()
	router.Handle("/health", handlers.NewHealthHandler()).Methods("GET")

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: router,
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start scheduler
	if err := schedulerService.Start(ctx); err != nil {
		logger.Fatal("Failed to start scheduler: %v", err)
	}

	// Start HTTP server
	go func() {
		logger.Info("Starting HTTP server on port %d", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start HTTP server: %v", err)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	// Shutdown gracefully
	logger.Info("Shutting down...")
	cancel()
	schedulerService.Stop()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Failed to shutdown HTTP server: %v", err)
	}
} 