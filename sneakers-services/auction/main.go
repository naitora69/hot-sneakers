package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"hotsneakers/auction/adapters/db"
	"hotsneakers/auction/config"
	"hotsneakers/closers"
)

func main() {
	// config
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()
	cfg := config.MustLoad(configPath)

	// logger
	log := mustMakeLogger(cfg.LogLevel)

	if err := run(cfg, log); err != nil {
		log.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	log.Info("starting server")
	log.Debug("debug messages are enabled")

	// database adapter
	pgRepo, err := db.New(log, cfg.DBAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %v", err)
	}
	if err := pgRepo.Migrate(); err != nil {
		return fmt.Errorf("failed to migrate db: %v", err)
	}
	defer closers.CloseOrLog(pgRepo.Conn, log)

	return nil
}

func mustMakeLogger(logLevel string) *slog.Logger {
	var level slog.Level
	switch logLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "ERROR":
		level = slog.LevelError
	default:
		panic("unknown log level: " + logLevel)
	}
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level})
	return slog.New(handler)
}
