package core

import (
	"context"
	"log/slog"
)

type Service struct {
	log *slog.Logger
	db  Repository
}

func NewService(log *slog.Logger, db Repository) (*Service, error) {
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

func (s Service) CreateSneaker(ctx context.Context, sneaker CreateSneaker) (int64, error) {
	id, err := s.db.CreateSneaker(ctx, sneaker)
	if err != nil {
		s.log.Error("error when create sneaker", "error", err)
		return -1, err
	}

	return id, nil
}

func (s Service) UpdateSneaker(ctx context.Context, sneaker UpdateSneaker) error {
	err := s.db.UpdateSneaker(ctx, sneaker)
	if err != nil {
		s.log.Error("error when update sneaker", "id", sneaker.ID, "error", err)
		return err
	}

	return nil
}
