package main

import (
	"context"
	"flag"
	"fmt"
	"hotsneakers/catalog/adapters/db"
	catalog "hotsneakers/catalog/adapters/grpc"
	"hotsneakers/catalog/config"
	"hotsneakers/catalog/core"
	"hotsneakers/closers"
	catalogpb "hotsneakers/proto/catalog"
	"log/slog"
	"net"
	"os"
	"os/signal"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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
	storage, err := db.New(log, cfg.DBAddress)
	if err != nil {
		return fmt.Errorf("failed to connect to db: %v", err)
	}
	if err := storage.Migrate(); err != nil {
		return fmt.Errorf("failed to migrate db: %v", err)
	}
	defer closers.CloseOrLog(storage.Conn, log)

	cataloger, err := core.NewService(log, storage)
	if err != nil {
		return fmt.Errorf("failed to create catalog service: %v", err)
	}

	// grpc server
	listener, err := net.Listen("tcp", cfg.CatalogAddress)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	catalogpb.RegisterCatalogServer(s, catalog.NewServer(cataloger))
	reflection.Register(s)

	// context for Ctrl-C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		s.GracefulStop()
	}()

	if err := s.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve: %v", err)
	}
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
