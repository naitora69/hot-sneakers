package core

import "context"

type DB interface {
	CreateAuction(ctx context.Context, auction Auction) (int, error)
	GetAuctionByID(ctx context.Context, id int) (Auction, error)
	PlaceBid(ctx context.Context, bid Bid) error
	GetBidsByAuctionID(ctx context.Context, auctionID int) ([]Bid, error)
}
type EventPublisher interface {
	PublishBidPlaced(ctx context.Context, bid Bid) error
}
type AuctionService interface {
	GetAuctionDetails(ctx context.Context, auctionID int) (Auction, []Bid, error)
	MakeBid(ctx context.Context, auctionID int, userID int, amount int64) error
	CreateAuction(ctx context.Context, auction Auction) (int, error)
}
