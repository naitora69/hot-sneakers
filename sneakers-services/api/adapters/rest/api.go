package rest

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"hotsneakers/api/core"
)

const (
	keyGetAuction = "auction_id"
	keySneakerID  = "id"
)

type PingResponse struct {
	Replies map[string]string `json:"replies"`
}

func NewPingHandler(log *slog.Logger, pingers map[string]core.Pinger) http.HandlerFunc {
	type pingResult struct {
		name   string
		status string
	}
	return func(w http.ResponseWriter, r *http.Request) {
		var wg sync.WaitGroup
		results := make(chan pingResult, len(pingers))
		replies := make(map[string]string, len(pingers))

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		for name, pinger := range pingers {
			wg.Go(func() {
				err := pinger.Ping(ctx)
				if err != nil {
					log.Warn("ping service error", "service", name, "error", err)
					results <- pingResult{
						name:   name,
						status: "unavailable",
					}
					return
				}

				results <- pingResult{
					name:   name,
					status: "ok",
				}
			})
		}

		go func() {
			wg.Wait()
			close(results)
		}()

		for res := range results {
			replies[res.name] = res.status
		}

		sendJSON(w, log, PingResponse{Replies: replies})
	}
}

func NewGetAuctionDetails(log *slog.Logger, auction core.AuctionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idString := r.URL.Query().Get(keyGetAuction)
		if idString == "" {
			http.Error(w, "auction_id is require field", http.StatusBadRequest)
			return
		}

		auctionId, err := strconv.Atoi(idString)
		if err != nil || auctionId <= 0 {
			log.Error("error when parse query to id", "error", err)
			http.Error(w, "auction_id must be integer greater then zero", http.StatusBadRequest)
			return
		}

		auction, bids, err := auction.GetAuctionDetails(r.Context(), auctionId)
		if err != nil {
			log.Error("error when get auction details", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		detailsDTO := struct {
			Auction core.Auction `json:"auction"`
			Bids    []core.Bid   `json:"bids"`
		}{
			Auction: auction,
			Bids:    bids,
		}

		sendJSON(w, log, detailsDTO)
	}
}

func NewCreateAuction(log *slog.Logger, auction core.AuctionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			SneakerID    int       `json:"sneaker_id"`
			CurrentPrice int64     `json:"start_price"`
			EndAt        time.Time `json:"end_at"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		newAuction := core.Auction{
			SneakerID:    req.SneakerID,
			CurrentPrice: req.CurrentPrice,
			EndAt:        req.EndAt,
		}

		id, err := auction.CreateAuction(r.Context(), newAuction)
		if err != nil {
			log.Error("failed to create auction", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		res := struct {
			AuctionID int `json:"auction_id"`
		}{
			AuctionID: id,
		}

		sendJSON(w, log, res)
	}
}

func NewMakeBid(log *slog.Logger, auction core.AuctionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			AuctionID int   `json:"auction_id"`
			UserID    int   `json:"user_id"`
			Amount    int64 `json:"amount"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		err := auction.MakeBid(r.Context(), req.AuctionID, req.UserID, req.Amount)
		if err != nil {
			log.Error("failed to make bid", "auction_id", req.AuctionID, "error", err)

			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}

		res := struct {
			Status string `json:"status"`
		}{
			Status: "bid_placed",
		}

		sendJSON(w, log, res)
	}
}

func NewGetAllSneakers(log *slog.Logger, catalog core.Cataloger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sneakers, err := catalog.GetAllSneakers(r.Context())
		if err != nil {
			log.Error("failed to get all sneakers", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		sendJSON(w, log, sneakers)
	}
}

func NewGetSneakerByID(log *slog.Logger, catalog core.Cataloger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idString := r.URL.Query().Get(keySneakerID)
		id, err := strconv.Atoi(idString)
		if err != nil || id <= 0 {
			http.Error(w, "id must be positive integer", http.StatusBadRequest)
			return
		}

		sneaker, err := catalog.GetSneakerByID(r.Context(), id)
		if err != nil {
			log.Error("failed to get sneaker", "id", id, "error", err)

			http.Error(w, "sneaker not found", http.StatusNotFound)
			return
		}

		sendJSON(w, log, sneaker)
	}
}

func NewCreateSneaker(log *slog.Logger, catalog core.Cataloger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req core.CreateSneaker
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		id, err := catalog.CreateSneaker(r.Context(), req)
		if err != nil {
			log.Error("failed to create sneaker", "error", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		res := struct {
			ID int64 `json:"id"`
		}{
			ID: id,
		}

		sendJSON(w, log, res)
	}
}

func sendJSON(w http.ResponseWriter, log *slog.Logger, data any) {
	body, err := json.Marshal(data)
	if err != nil {
		log.Error("failed to encode response", "error", err)
		http.Error(w, core.ErrInternal.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	_, err = w.Write(body)
	if err != nil {
		log.Error("write response error", "error", err)
		return
	}
}
