package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel    string `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	Address     string `yaml:"address" env:"AUCTION_ADDRESS" env-default:"localhost:80"`
	NatsAddress string `yaml:"nats_address" env:"NATS_ADDRESS" env-default:"localhost:4222"`
	DBAddress   string `yaml:"db_address" env:"DB_ADDRESS" env-default:"localhost:82"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
