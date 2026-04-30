package core

import "time"

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
