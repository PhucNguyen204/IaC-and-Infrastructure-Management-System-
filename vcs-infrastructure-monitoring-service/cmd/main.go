package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	httpHandler "github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/api/http"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/infrastructures/elasticsearch"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/infrastructures/kafka"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/collectors"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/env"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/pkg/logger"
	"github.com/PhucNguyen204/vcs-infrastructure-monitoring-service/usecases/services"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	envConfig, err := env.LoadEnv()
	if err != nil {
		log.Fatalf("Failed to load env: %v", err)
	}

	logger, err := logger.LoadLogger(envConfig.LoggerEnv)
	if err != nil {
		log.Fatalf("Failed to init logger: %v", err)
	}
	defer logger.Sync()

	redisClient := redis.NewClient(&redis.Options{
		Addr:     envConfig.RedisEnv.RedisHost + ":" + envConfig.RedisEnv.RedisPort,
		Password: envConfig.RedisEnv.RedisPassword,
		DB:       envConfig.RedisEnv.RedisDB,
	})
	defer redisClient.Close()

	esClient, err := elasticsearch.NewElasticsearchClient(envConfig.ElasticsearchEnv, logger)
	if err != nil {
		log.Fatalf("Failed to create elasticsearch client: %v", err)
	}

	kafkaConsumer := kafka.NewKafkaConsumer(envConfig.KafkaEnv, esClient, logger)
	defer kafkaConsumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := kafkaConsumer.Start(ctx); err != nil {
		log.Fatalf("Failed to start kafka consumer: %v", err)
	}

	var dockerCollector collectors.IDockerCollector
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logger.Warn("Failed to create docker client, health checks will be limited", zap.Error(err))
		dockerCollector = nil
	} else {
		defer dockerClient.Close()
		dockerCollector = collectors.NewDockerCollector(dockerClient, logger)
	}

	healthCheckService := services.NewHealthCheckService(dockerCollector, esClient, redisClient, logger)
	if err := healthCheckService.Start(ctx); err != nil {
		log.Fatalf("Failed to start health check service: %v", err)
	}

	metricsService := services.NewMetricsService(esClient, logger)
	uptimeService := services.NewUptimeService(esClient, logger)

	monitoringHandler := httpHandler.NewMonitoringHandler(metricsService, redisClient)
	uptimeHandler := httpHandler.NewUptimeHandler(uptimeService)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	apiV1 := r.Group("/api/v1")
	monitoringHandler.RegisterRoutes(apiV1)
	uptimeHandler.RegisterRoutes(apiV1)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr:    ":" + envConfig.HTTPEnv.Port,
		Handler: r,
	}

	go func() {
		<-quit
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP server shutdown failed", zap.Error(err))
		}
		logger.Info("Infrastructure monitoring service stopped gracefully")
	}()

	logger.Info("Infrastructure monitoring service started", zap.String("port", envConfig.HTTPEnv.Port))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to run service: %v", err)
	}
}
