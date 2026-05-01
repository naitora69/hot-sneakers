package core

import "context"

type Pinger interface {
	Ping(context.Context) error
}

type Cataloger interface {
	GetAllSneakers(context.Context) ([]Sneaker, error)
	GetSneakerByID(context.Context, int) (Sneaker, error)
	CreateSneaker(context.Context, CreateSneaker) (int64, error)
	UpdateSneaker(context.Context, UpdateSneaker) error
}
