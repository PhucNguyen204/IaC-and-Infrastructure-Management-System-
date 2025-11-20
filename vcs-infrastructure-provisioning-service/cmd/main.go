package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	httpHandler "github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/api/http"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/entities"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/databases"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/docker"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/infrastructures/kafka"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/env"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/middlewares"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/repositories"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/usecases/services"
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

	postgresDb, err := databases.ConnectPostgresDb(envConfig.PostgresEnv)
	if err != nil {
		log.Fatalf("Failed to connect to postgres: %v", err)
	}

	if err := postgresDb.AutoMigrate(
		&entities.Infrastructure{},
		&entities.PostgreSQLInstance{},
		&entities.NginxInstance{},
		&entities.PostgreSQLCluster{},
		&entities.ClusterNode{},
		&entities.EtcdNode{},
		&entities.NginxDomain{},
		&entities.NginxRoute{},
		&entities.NginxUpstream{},
		&entities.NginxUpstreamBackend{},
		&entities.NginxCertificate{},
		&entities.NginxSecurity{},
		&entities.PostgresDatabase{},
		&entities.PostgresBackup{},
		&entities.DockerService{},
		&entities.DockerEnvVar{},
		&entities.DockerPort{},
		&entities.DockerNetwork{},
		&entities.DockerHealthCheck{},
		&entities.Stack{},
		&entities.StackResource{},
		&entities.StackTemplate{},
		&entities.StackOperation{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	dockerService, err := docker.NewDockerService(logger)
	if err != nil {
		log.Fatalf("Failed to create docker service: %v", err)
	}

	kafkaProducer := kafka.NewKafkaProducer(envConfig.KafkaEnv, logger)
	defer kafkaProducer.Close()

	infraRepo := repositories.NewInfrastructureRepository(postgresDb)
	pgRepo := repositories.NewPostgreSQLRepository(postgresDb)
	nginxRepo := repositories.NewNginxRepository(postgresDb)
	clusterRepo := repositories.NewPostgreSQLClusterRepository(postgresDb)
	pgDatabaseRepo := repositories.NewPostgresDatabaseRepository(postgresDb)
	dockerRepo := repositories.NewDockerServiceRepository(postgresDb)
	stackRepo := repositories.NewStackRepository(postgresDb)

	pgService := services.NewPostgreSQLService(infraRepo, pgRepo, dockerService, kafkaProducer, logger)
	nginxService := services.NewNginxService(infraRepo, nginxRepo, dockerService, kafkaProducer, logger)
	clusterService := services.NewPostgreSQLClusterService(infraRepo, clusterRepo, dockerService, kafkaProducer, logger)
	pgDatabaseService := services.NewPostgresDatabaseService(pgDatabaseRepo, pgRepo, dockerService)
	dockerSvcService := services.NewDockerServiceService(dockerRepo, infraRepo, dockerService)
	stackService := services.NewStackService(stackRepo, infraRepo, nginxService, pgService, clusterService, pgDatabaseService, dockerSvcService)

	jwtMiddleware := middlewares.NewJWTMiddleware(envConfig.AuthEnv.JWTSecret)

	pgHandler := httpHandler.NewPostgreSQLHandler(pgService)
	nginxHandler := httpHandler.NewNginxHandler(nginxService)
	clusterHandler := httpHandler.NewPostgreSQLClusterHandler(clusterService, logger)
	pgDatabaseHandler := httpHandler.NewPostgresDatabaseHandler(pgDatabaseService)
	dockerHandler := httpHandler.NewDockerServiceHandler(dockerSvcService)
	stackHandler := httpHandler.NewStackHandler(stackService)

	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	apiV1 := r.Group("/api/v1", jwtMiddleware.CheckBearerAuth())
	pgHandler.RegisterRoutes(apiV1)
	nginxHandler.RegisterRoutes(apiV1)
	pgDatabaseHandler.RegisterRoutes(apiV1)
	dockerHandler.RegisterRoutes(apiV1)
	stackHandler.RegisterRoutes(apiV1)

	// PostgreSQL Cluster routes
	clusterGroup := apiV1.Group("/postgres/cluster")
	{
		clusterGroup.POST("", clusterHandler.CreateCluster)
		clusterGroup.GET("/:id", clusterHandler.GetClusterInfo)
		clusterGroup.POST("/:id/start", clusterHandler.StartCluster)
		clusterGroup.POST("/:id/stop", clusterHandler.StopCluster)
		clusterGroup.POST("/:id/restart", clusterHandler.RestartCluster)
		clusterGroup.DELETE("/:id", clusterHandler.DeleteCluster)
		clusterGroup.POST("/:id/scale", clusterHandler.ScaleCluster)
		clusterGroup.POST("/:id/failover", clusterHandler.PromoteReplica)
		clusterGroup.GET("/:id/replication", clusterHandler.GetReplicationStatus)
		clusterGroup.GET("/:id/stats", clusterHandler.GetClusterStats)
		clusterGroup.GET("/:id/logs", clusterHandler.GetClusterLogs)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{
		Addr:    ":" + envConfig.HTTPEnv.Port,
		Handler: r,
	}

	go func() {
		<-quit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			logger.Error("HTTP server shutdown failed", zap.Error(err))
		}
		logger.Info("Infrastructure provisioning service stopped gracefully")
	}()

	logger.Info("Infrastructure provisioning service started", zap.String("port", envConfig.HTTPEnv.Port))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to run service: %v", err)
	}
}
