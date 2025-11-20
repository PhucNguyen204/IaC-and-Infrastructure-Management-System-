package databases

import (
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/PhucNguyen204/vcs-infrastructure-provisioning-service/pkg/env"
)

type RedisFactory struct {
	env env.RedisEnv
}

func NewRedisFactory(env env.RedisEnv) *RedisFactory {
	return &RedisFactory{env: env}
}

func (rf *RedisFactory) ConnectRedis() *redis.Client {
	addr := fmt.Sprintf("%s:%s", rf.env.RedisHost, rf.env.RedisPort)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: rf.env.RedisPassword,
		DB:       rf.env.RedisDB,
	})
	return client
}
