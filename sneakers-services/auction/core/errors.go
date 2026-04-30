package core

import (
	"errors"
)

var (
	ErrAuctionIsEnded         = errors.New("auction is ended")
	ErrBidToLow               = errors.New("your bid is too low")
	ErrAuctionNotFound        = errors.New("auction not found")
	ErrInvalidAuctionOrUserId = errors.New("invalid auction or user id")
)
