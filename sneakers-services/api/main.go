package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"

	"hotsneakers/api/adapters/auction"
	"hotsneakers/api/adapters/catalog"
	"hotsneakers/api/adapters/rest"
	"hotsneakers/api/config"
	"hotsneakers/api/core"
	"hotsneakers/closers"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "config.yaml", "server configuration file")
	flag.Parse()

	cfg := config.MustLoad(configPath)
	log := mustMakeLogger(cfg.LogLevel)

	if err := run(&cfg, log); err != nil {
		log.Error("server stopped with error", "error", err)
		os.Exit(1)
	}
}

func run(cfg *config.Config, log *slog.Logger) error {
	log.Info("starting server")
	log.Debug("debug messages are enabled")

	catalogClient, err := catalog.NewClient(cfg.CatalogAddress, log)
	if err != nil {
		return fmt.Errorf("cannot init catalog adapter: %w", err)
	}
	defer closers.CloseOrLog(catalogClient.Conn, log)

	auctionClient, err := auction.NewClient(cfg.AuctionAddress, log)
	if err != nil {
		return fmt.Errorf("cannot init auction adapter: %w", err)
	}
	defer closers.CloseOrLog(auctionClient.Conn, log)

	mux := http.NewServeMux()

	mux.Handle("GET /api/ping", rest.NewPingHandler(log, map[string]core.Pinger{"catalog": catalogClient, "auction": auctionClient}))
	mux.Handle("GET /api/auction", rest.NewGetAuctionDetails(log, auctionClient))
	mux.Handle("POST /api/auction/create", rest.NewCreateAuction(log, auctionClient))
	mux.Handle("POST /api/auction/bid", rest.NewMakeBid(log, auctionClient))
	mux.Handle("GET /api/sneakers", rest.NewGetAllSneakers(log, catalogClient))
	mux.Handle("GET /api/sneaker", rest.NewGetSneakerByID(log, catalogClient))
	mux.Handle("POST /api/sneaker/create", rest.NewCreateSneaker(log, catalogClient))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	server := http.Server{
		Addr:        cfg.HTTPConfig.Address,
		ReadTimeout: cfg.HTTPConfig.Timeout,
		Handler:     mux,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
	}

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("erroneous shutdown", "error", err)
		}
	}()

	log.Info("Running HTTP server", "address", cfg.HTTPConfig.Address)
	if err := server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server closed unexpectedly: %w", err)
		}
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
