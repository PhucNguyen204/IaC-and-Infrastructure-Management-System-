package env

import (
	"github.com/spf13/viper"
)

type Env struct {
	PostgresEnv PostgresEnv
	RedisEnv    RedisEnv
	KafkaEnv    KafkaEnv
	LoggerEnv   LoggerEnv
	GRPCEnv     GRPCEnv
	HTTPEnv     HTTPEnv
	AuthEnv     AuthEnv
}

type AuthEnv struct {
	JWTSecret string
}

type PostgresEnv struct {
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresName     string
}

type RedisEnv struct {
	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int
}

type KafkaEnv struct {
	Brokers []string
	Topic   string
}

type LoggerEnv struct {
	Level      string
	FilePath   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
}

type GRPCEnv struct {
	Port string
}

type HTTPEnv struct {
	Port string
}

func LoadEnv() (*Env, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	viper.ReadInConfig()

	return &Env{
		PostgresEnv: PostgresEnv{
			PostgresHost:     viper.GetString("POSTGRES_HOST"),
			PostgresPort:     viper.GetString("POSTGRES_PORT"),
			PostgresUser:     viper.GetString("POSTGRES_USER"),
			PostgresPassword: viper.GetString("POSTGRES_PASSWORD"),
			PostgresName:     viper.GetString("POSTGRES_DB"),
		},
		RedisEnv: RedisEnv{
			RedisHost:     viper.GetString("REDIS_HOST"),
			RedisPort:     viper.GetString("REDIS_PORT"),
			RedisPassword: viper.GetString("REDIS_PASSWORD"),
			RedisDB:       viper.GetInt("REDIS_DB"),
		},
		KafkaEnv: KafkaEnv{
			Brokers: viper.GetStringSlice("KAFKA_BROKERS"),
			Topic:   viper.GetString("KAFKA_TOPIC"),
		},
		LoggerEnv: LoggerEnv{
			Level:      viper.GetString("LOG_LEVEL"),
			FilePath:   viper.GetString("LOG_FILE_PATH"),
			MaxSize:    viper.GetInt("LOG_MAX_SIZE"),
			MaxAge:     viper.GetInt("LOG_MAX_AGE"),
			MaxBackups: viper.GetInt("LOG_MAX_BACKUPS"),
		},
		GRPCEnv: GRPCEnv{
			Port: viper.GetString("GRPC_PORT"),
		},
		HTTPEnv: HTTPEnv{
			Port: viper.GetString("HTTP_PORT"),
		},
		AuthEnv: AuthEnv{
			JWTSecret: viper.GetString("JWT_SECRET"),
		},
	}, nil
}

