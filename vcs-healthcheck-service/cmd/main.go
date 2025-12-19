package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/PhucNguyen204/vcs-healthcheck-service/api/http"
	"github.com/PhucNguyen204/vcs-healthcheck-service/infrastructures/docker"
	"github.com/PhucNguyen204/vcs-healthcheck-service/infrastructures/elasticsearch"
	"github.com/PhucNguyen204/vcs-healthcheck-service/infrastructures/kafka"
	"github.com/PhucNguyen204/vcs-healthcheck-service/pkg/env"
	"github.com/PhucNguyen204/vcs-healthcheck-service/pkg/logger"
	"github.com/PhucNguyen204/vcs-healthcheck-service/usecases/services"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	config := env.LoadConfig()

	// Initialize logger
	log := logger.NewLogger(logger.LoggerConfig{
		Level:      config.Logger.Level,
		FilePath:   config.Logger.FilePath,
		MaxSize:    config.Logger.MaxSize,
		MaxAge:     config.Logger.MaxAge,
		MaxBackups: config.Logger.MaxBackups,
	})
	defer log.Sync()

	log.Info("starting healthcheck service")

	// Initialize Kafka consumer
	kafkaConsumer := kafka.NewKafkaConsumer(kafka.ConsumerConfig{
		Brokers:       config.Kafka.Brokers,
		ConsumerGroup: config.Kafka.ConsumerGroup,
		Topics:        config.Kafka.Topics,
	}, log)

	// Initialize Elasticsearch client
	esClient, err := elasticsearch.NewElasticsearchClient(elasticsearch.Config{
		Addresses: config.Elasticsearch.Addresses,
		Username:  config.Elasticsearch.Username,
		Password:  config.Elasticsearch.Password,
	}, log)
	if err != nil {
		log.Fatal("failed to create elasticsearch client", zap.Error(err))
	}

	// Initialize Docker client
	dockerClient, err := docker.NewDockerClient(log)
	if err != nil {
		log.Fatal("failed to create docker client", zap.Error(err))
	}

	// Initialize health check service
	healthCheckService := services.NewHealthCheckService(
		kafkaConsumer,
		esClient,
		dockerClient,
		log,
	)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start health check service
	if err := healthCheckService.Start(ctx); err != nil {
		log.Fatal("failed to start health check service", zap.Error(err))
	}

	// Setup HTTP server
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())

	// Register routes
	api := router.Group("/api/v1")
	healthHandler := http.NewHealthHandler(log)
	healthHandler.RegisterRoutes(api)

	// Start HTTP server in goroutine
	go func() {
		addr := fmt.Sprintf(":%d", config.Server.Port)
		log.Info("starting HTTP server", zap.String("address", addr))
		if err := router.Run(addr); err != nil {
			log.Error("http server error", zap.Error(err))
		}
	}()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("shutting down...")
	cancel()
	healthCheckService.Stop()
	log.Info("shutdown complete")
}
