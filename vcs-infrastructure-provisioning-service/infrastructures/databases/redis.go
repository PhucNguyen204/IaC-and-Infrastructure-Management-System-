package databases

import (
	"context"
	"fmt"
	"time"

	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/env"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisFactory struct {
	env    env.RedisEnv
	logger logger.ILogger
}

func NewRedisFactory(env env.RedisEnv, logger logger.ILogger) *RedisFactory {
	return &RedisFactory{env: env, logger: logger}
}

func (rf *RedisFactory) ConnectRedis() *redis.Client {
	// Check if Redis is configured
	if rf.env.RedisHost == "" {
		rf.logger.Warn("Redis not configured, caching disabled")
		return nil
	}

	addr := fmt.Sprintf("%s:%s", rf.env.RedisHost, rf.env.RedisPort)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: rf.env.RedisPassword,
		DB:       rf.env.RedisDB,
	})

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		rf.logger.Warn("Failed to connect to Redis, caching disabled",
			zap.String("addr", addr),
			zap.Error(err))
		return nil
	}

	rf.logger.Info("Connected to Redis", zap.String("addr", addr))
	return client
}
