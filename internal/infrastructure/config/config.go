package config

import (
	"github.com/spf13/viper"
)

type CacheConfig struct {
	Capacity    int `mapstructure:"capacity"`
	GetAllLimit int `mapstructure:"get_all_limit"`
}

type MigrationsConfig struct {
	MigrationsPath string `mapstructure:"migrations_path"`
}

type Config struct {
	Cache      CacheConfig      `mapstructure:"cache"`
	Migrations MigrationsConfig `mapstructure:"migrations"`
}

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
