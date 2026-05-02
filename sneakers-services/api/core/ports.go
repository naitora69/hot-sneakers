package core

import "context"

type Pinger interface {
	Ping(context.Context) error
}

type Cataloger interface {
	GetAllSneakers(ctx context.Context) ([]Sneaker, error)
	GetSneakerByID(ctx context.Context, is int) (Sneaker, error)
	CreateSneaker(ctx context.Context, sneaker CreateSneaker) (int64, error)
	UpdateSneaker(ctx context.Context, sneaker UpdateSneaker) error
}

type AuctionService interface {
	GetAuctionDetails(ctx context.Context, auctionID int) (Auction, []Bid, error)
	MakeBid(ctx context.Context, auctionID int, userID int, amount int64) error
	CreateAuction(ctx context.Context, auction Auction) (int, error)
}
