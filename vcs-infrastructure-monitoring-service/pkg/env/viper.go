package env

import (
	"github.com/spf13/viper"
)

type Env struct {
	PostgresEnv       PostgresEnv
	RedisEnv          RedisEnv
	KafkaEnv          KafkaEnv
	ElasticsearchEnv  ElasticsearchEnv
	LoggerEnv         LoggerEnv
	HTTPEnv           HTTPEnv
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
	GroupID string
	Topics  []string
}

type ElasticsearchEnv struct {
	Addresses []string
	Username  string
	Password  string
}

type LoggerEnv struct {
	Level      string
	FilePath   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
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
			GroupID: viper.GetString("KAFKA_GROUP_ID"),
			Topics:  viper.GetStringSlice("KAFKA_TOPICS"),
		},
		ElasticsearchEnv: ElasticsearchEnv{
			Addresses: viper.GetStringSlice("ELASTICSEARCH_ADDRESSES"),
			Username:  viper.GetString("ELASTICSEARCH_USERNAME"),
			Password:  viper.GetString("ELASTICSEARCH_PASSWORD"),
		},
		LoggerEnv: LoggerEnv{
			Level:      viper.GetString("LOG_LEVEL"),
			FilePath:   viper.GetString("LOG_FILE_PATH"),
			MaxSize:    viper.GetInt("LOG_MAX_SIZE"),
			MaxAge:     viper.GetInt("LOG_MAX_AGE"),
			MaxBackups: viper.GetInt("LOG_MAX_BACKUPS"),
		},
		HTTPEnv: HTTPEnv{
			Port: viper.GetString("HTTP_PORT"),
		},
	}, nil
}

