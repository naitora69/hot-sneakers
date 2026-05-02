package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"hotsneakers/notification/config"

	"github.com/gorilla/websocket"
	"github.com/nats-io/nats.go"
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
		return
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func newWebSocketHandler(nc *nats.Conn, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auctionID := r.URL.Query().Get("auction_id")
		if auctionID == "" {
			http.Error(w, "auction_id is required field", http.StatusBadRequest)
			return
		}

		ws, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error("error when upgrade websocket", "error", err)
			return
		}
		defer func() {
			err := ws.Close()
			if err != nil {
				log.Error("fatal error when close websocket", "error", err)
			}
		}()

		writeChan := make(chan []byte, 10)
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		subject := "auction.bids." + auctionID

		sub, err := nc.Subscribe(subject, func(msg *nats.Msg) {
			select {
			case writeChan <- msg.Data:
			case <-ctx.Done():
			}
		})
		if err != nil {
			log.Error("nats subscribe error", "error", err)
			return
		}

		defer func() {
			err := sub.Unsubscribe()
			log.Error("fatal error when unsubscribe", "error", err)
		}()

		go func() {
			for data := range writeChan {
				if err := ws.WriteMessage(websocket.TextMessage, data); err != nil {
					log.Error("ws write error", "error", err)
					cancel()
					return
				}
			}
		}()

		for {
			_, _, err := ws.ReadMessage()
			if err != nil {
				log.Info("client disconnected")
				break
			}
		}
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	nc, err := nats.Connect(cfg.NatsAddress)
	if err != nil {
		return err
	}
	defer nc.Close()

	mux := http.NewServeMux()
	mux.Handle("GET /ws/auction", newWebSocketHandler(nc, log))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	server := http.Server{
		Addr:    cfg.WSAddress,
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		log.Debug("shutting down server")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Error("erroneous shutdown", "error", err)
		}
	}()

	log.Info("Notification service is running", "address", cfg.WSAddress)
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
