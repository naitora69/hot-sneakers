package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	LogLevel       string `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	CatalogAddress string `yaml:"catalog_address" env:"CATALOG_ADDRESS" end-default:"localhost:80"`
	DBAddress      string `yaml:"db_address" env:"DB_ADDRESS" env-default:"localhost:82"`
	RedisAddress   string `yaml:"redis_address" env:"REDIS_ADDRESS" env-default:"localhost:6379"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
