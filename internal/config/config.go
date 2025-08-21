package config

import (
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	TimeDurationPublisher time.Duration `yaml:"time_duration_publisher"`
	Port int `yaml:"port"`
}

func MustLoad() (*Config, error) {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "./config/local.yaml"
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, err
	}
	var config Config	
	if err := cleanenv.ReadConfig(configPath, &config); err != nil {
		return nil, err
	}
	return &config, nil
}
