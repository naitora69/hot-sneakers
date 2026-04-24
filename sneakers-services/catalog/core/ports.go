package core

import (
	"context"
)

type Cataloger interface {
	GetAllSneakers(context.Context) ([]Sneaker, error)
	GetSneakerByID(context.Context, int) (Sneaker, error)
}

type DB interface {
	GetAllSneakers(context.Context) ([]Sneaker, error)
	GetSneakerByID(context.Context, int) (Sneaker, error)
}
