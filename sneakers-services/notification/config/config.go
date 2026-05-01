package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel    string `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	NatsAddress string `yaml:"nats_address" env:"NATS_ADDRESS" env-default:"localhost:4222"`
	WSAddress   string `yaml:"websocket_address" env:"WEBSOCKET_ADDRESS" env-default:"localhost:81"`
	DBAddress   string `yaml:"db_address" env:"DB_ADDRESS" env-default:"localhost:82"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
