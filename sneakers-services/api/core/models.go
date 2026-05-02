package core

import "time"

type Sneaker struct {
	ID    int
	Brand string
	Model string
}

type CreateSneaker struct {
	Brand string
	Model string
}

type UpdateSneaker struct {
	ID    int64
	Brand string
	Model string
}

type Auction struct {
	ID           int
	SneakerID    int
	CurrentPrice int64
	EndAt        time.Time
}

type Bid struct {
	ID        int
	AuctionID int
	UserID    int
	Amount    int64
}
