package db

import (
	"context"
	"fmt"
	"hotsneakers/catalog/core"
	"log/slog"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type DB struct {
	log  *slog.Logger
	Conn *sqlx.DB
}
type sneakerDTO struct {
	ID    int    `db:"id"`
	Brand string `db:"brand"`
	Model string `db:"model"`
}

func New(log *slog.Logger, address string) (*DB, error) {
	db, err := sqlx.Connect("pgx", address)
	if err != nil {
		log.Error("connection problem", "address", address, "error", err)
		return nil, err
	}

	return &DB{
		log:  log,
		Conn: db,
	}, nil
}

func (db DB) GetAllSneakers(ctx context.Context) ([]core.Sneaker, error) {
	query := `SELECT id, brand, model FROM sneakers;`

	sds := []sneakerDTO{}

	err := db.Conn.SelectContext(ctx, &sds, query)
	if err != nil {
		return nil, fmt.Errorf("db adapter error: %w", err)
	}

	res := make([]core.Sneaker, 0, len(sds))
	for _, sd := range sds {
		tmpSneaker := core.Sneaker(sd)
		res = append(res, tmpSneaker)
	}

	return res, nil
}

func (db DB) GetSneakerByID(ctx context.Context, id int) (core.Sneaker, error) {
	query := `SELECT brand, model FROM sneakers WHERE id=$1;`

	sd := sneakerDTO{}

	err := db.Conn.GetContext(ctx, &sd, query, id)

	if err != nil {
		return core.Sneaker{}, fmt.Errorf("dp adapter error: %w", err)
	}

	res := core.Sneaker(sd)

	return res, nil
}
