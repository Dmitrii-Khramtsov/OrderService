// github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config/config.go
package config

import (
	"time"

	"github.com/spf13/viper"
)

type RestorationConfig struct {
	Timeout   int `mapstructure:"timeout"`
	BatchSize int `mapstructure:"batch_size"`
}

type CacheConfig struct {
	Capacity    int               `mapstructure:"capacity"`
	GetAllLimit int               `mapstructure:"get_all_limit"`
	Restoration RestorationConfig `mapstructure:"restoration"`
}

type DatabaseConfig struct {
	DSN             string        `mapstructure:"dsn"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type KafkaConfig struct {
	Brokers        []string      `mapstructure:"brokers"`
	Topic          string        `mapstructure:"topic"`
	GroupID        string        `mapstructure:"group_id"`
	DLQTopic       string        `mapstructure:"dlq_topic"`
	MaxRetries     int           `mapstructure:"max_retries"`
	ProcessingTime time.Duration `yaml:"processing_time"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type MigrationsConfig struct {
	MigrationsPath string `mapstructure:"migrations_path"`
}

type RetryConfig struct {
	MaxElapsedTime      time.Duration `mapstructure:"max_elapsed_time"`
	InitialInterval     time.Duration `mapstructure:"initial_interval"`
	RandomizationFactor float64       `mapstructure:"randomization_factor"`
	Multiplier          float64       `mapstructure:"multiplier"`
	MaxInterval         time.Duration `mapstructure:"max_interval"`
}

type Config struct {
	Cache      CacheConfig      `mapstructure:"cache"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Kafka      KafkaConfig      `mapstructure:"kafka"`
	Server     ServerConfig     `mapstructure:"server"`
	Migrations MigrationsConfig `mapstructure:"migrations"`
	Retry      RetryConfig      `mapstructure:"retry"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	viper.BindEnv("database.dsn", "POSTGRES_DSN")
	viper.BindEnv("kafka.brokers", "KAFKA_BROKERS")
	viper.BindEnv("log.mode", "LOG_MODE")

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
