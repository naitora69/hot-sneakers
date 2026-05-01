package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPConfig struct {
	Address string        `yaml:"address" env:"API_ADDRESS" env-default:"localhost:80"`
	Timeout time.Duration `yaml:"timeout" env:"API_TIMEOUT" env-default:"15s"`
}

type Config struct {
	LogLevel       string     `yaml:"log_level" env:"LOG_LEVEL" env-default:"DEBUG"`
	HTTPConfig     HTTPConfig `yaml:"api_server"`
	CatalogAddress string     `yaml:"catalog_address" env:"CATALOG_ADDRESS" end-default:"localhost:81"`
	AuctionAddress string     `yaml:"auction_address" env:"AUCTION_ADDRESS" end-default:"localhost:82"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
