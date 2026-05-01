package core

import (
	"context"
	"log/slog"
)

type Service struct {
	log    *slog.Logger
	db     DB
	broker EventPublisher
}

func NewService(log *slog.Logger, db DB, broker EventPublisher) *Service {
	return &Service{
		log:    log,
		db:     db,
		broker: broker,
	}
}

func (s Service) GetAuctionDetails(ctx context.Context, auctionID int) (Auction, []Bid, error) {
	auction, err := s.db.GetAuctionByID(ctx, auctionID)
	if err != nil {
		return Auction{}, []Bid{}, err
	}

	bids, err := s.db.GetBidsByAuctionID(ctx, auctionID)
	if err != nil {
		return Auction{}, []Bid{}, err
	}

	return auction, bids, nil
}

func (s Service) MakeBid(ctx context.Context, auctionID int, userID int, amount int64) error {
	if auctionID <= 0 || userID <= 0 {
		return ErrInvalidAuctionOrUserId
	}

	bidToPlace := Bid{AuctionID: auctionID, UserID: userID, Amount: amount}

	err := s.db.PlaceBid(ctx, bidToPlace)
	if err != nil {
		return err
	}

	return nil
}

func (s Service) CreateAuction(ctx context.Context, auction Auction) (int, error) {
	id, err := s.db.CreateAuction(ctx, auction)
	if err != nil {
		return 0, err
	}

	return id, nil
}
