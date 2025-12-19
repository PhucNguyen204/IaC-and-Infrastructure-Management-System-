package env

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server        ServerConfig
	Kafka         KafkaConfig
	Elasticsearch ElasticsearchConfig
	Redis         RedisConfig
	Docker        DockerConfig
	Logger        LoggerConfig
}

type ServerConfig struct {
	Port int
}

type KafkaConfig struct {
	Brokers       []string
	ConsumerGroup string
	Topics        []string
}

type ElasticsearchConfig struct {
	Addresses []string
	Username  string
	Password  string
}

type RedisConfig struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type DockerConfig struct {
	Host string
}

type LoggerConfig struct {
	Level      string
	FilePath   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
}

func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnvInt("HTTP_PORT", 8085),
		},
		Kafka: KafkaConfig{
			Brokers:       getEnvSlice("KAFKA_BROKERS", []string{"localhost:9092"}),
			ConsumerGroup: getEnv("KAFKA_CONSUMER_GROUP", "healthcheck-service"),
			Topics:        getEnvSlice("KAFKA_TOPICS", []string{"infrastructure.lifecycle"}),
		},
		Elasticsearch: ElasticsearchConfig{
			Addresses: getEnvSlice("ES_ADDRESSES", []string{"http://localhost:9200"}),
			Username:  getEnv("ES_USERNAME", ""),
			Password:  getEnv("ES_PASSWORD", ""),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnvInt("REDIS_PORT", 6379),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		Docker: DockerConfig{
			Host: getEnv("DOCKER_HOST", "unix:///var/run/docker.sock"),
		},
		Logger: LoggerConfig{
			Level:      getEnv("LOG_LEVEL", "info"),
			FilePath:   getEnv("LOG_FILE_PATH", "./logs/healthcheck.log"),
			MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
			MaxAge:     getEnvInt("LOG_MAX_AGE", 10),
			MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 30),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
