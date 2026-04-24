package core

import (
	"context"
	"log/slog"
)

type Service struct {
	log *slog.Logger
	db  DB
}

func NewService(log *slog.Logger, db DB) (*Service, error) {
	service := &Service{
		log: log,
		db:  db,
	}

	return service, nil
}
func (s Service) GetAllSneakers(ctx context.Context) ([]Sneaker, error) {
	sneakers, err := s.db.GetAllSneakers(ctx)
	if err != nil {
		s.log.Error("error when get all sneakers", "error", err)
		return nil, err
	}

	return sneakers, nil
}
func (s Service) GetSneakerByID(ctx context.Context, id int) (Sneaker, error) {
	sneaker, err := s.db.GetSneakerByID(ctx, id)
	if err != nil {
		s.log.Error("error when get sneaker by id", "error", err)
		return Sneaker{}, err
	}

	return sneaker, nil
}
