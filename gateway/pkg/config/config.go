package config

import (
	"log/slog"

	"github.com/spf13/viper"
)

type Config struct {
	// Server
	ServerPort string `mapstructure:"SERVER_PORT"`

	// Database (PostgreSQL)
	PostgresHost string `mapstructure:"POSTGRES_HOST"`
	PostgresPort string `mapstructure:"POSTGRES_PORT"`
	PostgresUser string `mapstructure:"POSTGRES_USER"`
	PostgresPass string `mapstructure:"POSTGRES_PASSWORD"`
	PostgresName string `mapstructure:"POSTGRES_DB"`

	// Cache (Redis)
	RedisHost string `mapstructure:"REDIS_HOST"`
	RedisPort string `mapstructure:"REDIS_PORT"`

	// Object Storage (MinIO)
	MinioHost      string `mapstructure:"MINIO_HOST"`
	MinioPort      string `mapstructure:"MINIO_PORT"`
	MinioPanelPort string `mapstructure:"MINIO_PANEL_PORT"`
	MinioAccessKey string `mapstructure:"MINIO_ROOT_USER"`
	MinioSecretKey string `mapstructure:"MINIO_ROOT_PASSWORD"`

	// Logger
	LogLevel      string `mapstructure:"LOG_LEVEL"`
	ConsoleOutput string `mapstructure:"CONSOLE_OUTPUT"`
}

func LoadConfig() (config Config, err error) {
	viper.AddConfigPath("./")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")

	viper.AutomaticEnv()

	if err = viper.ReadInConfig(); err != nil {
		slog.Error("error reading config file", "error", err)
	}

	if err = viper.Unmarshal(&config); err != nil {
		slog.Error("error unmarshaling config", "error", err)
	}

	return
}
