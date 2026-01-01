package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hr-management-system/internal/config"
	"hr-management-system/internal/delivery/http/router"
	"hr-management-system/internal/i18n"
	"hr-management-system/internal/infrastructure/cache"
	"hr-management-system/internal/infrastructure/database"
	"hr-management-system/internal/infrastructure/email"
	"hr-management-system/internal/infrastructure/logger"
	"hr-management-system/internal/infrastructure/queue"
	"hr-management-system/internal/infrastructure/search"
	"hr-management-system/internal/security"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err := logger.NewLogger(&cfg.Logger)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Close()

	log.Info("Starting HR Management System API...")

	_, err = i18n.New("vi")
	if err != nil {
		log.WithError(err).Error("Failed to initialize i18n")
		os.Exit(1)
	}

	security.Init(&cfg.Security)

	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()
	log.Info("Connected to PostgreSQL")

	redisCache, err := cache.NewRedisCache(&cfg.Redis)
	if err != nil {
		log.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisCache.Close()
	log.Info("Connected to Redis")

	es, err := search.NewElasticSearch(&cfg.Elastic)
	if err != nil {
		log.WithError(err).Warn("Failed to connect to Elasticsearch, search features will be limited")
	} else {
		if err := es.InitIndices(context.Background()); err != nil {
			log.WithError(err).Warn("Failed to initialize Elasticsearch indices")
		}
		log.Info("Connected to Elasticsearch")
	}

	jobQueue, err := queue.NewQueue(&cfg.Worker)
	if err != nil {
		log.WithError(err).Warn("Failed to initialize job queue")
	} else {
		defer jobQueue.Close()
		log.Info("Job queue initialized")
	}

	emailSvc, err := email.NewEmailService(&cfg.Email)
	if err != nil {
		log.WithError(err).Warn("Failed to initialize email service")
	} else {
		log.Info("Email service initialized")
	}

	r := router.NewRouter(cfg, db, redisCache, jobQueue, es, emailSvc, log)
	engine := r.Setup()

	server := &http.Server{
		Addr:         ":" + cfg.App.Port,
		Handler:      engine,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.WithField("port", cfg.App.Port).Info("Server starting...")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("Server failed to start")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("Server forced to shutdown")
	}

	log.Info("Server stopped")
}
