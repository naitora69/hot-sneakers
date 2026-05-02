package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"hotsneakers/auction/core"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type auctionDTO struct {
	ID           int       `db:"id"`
	SneakerID    int       `db:"sneaker_id"`
	CurrentPrice int64     `db:"current_price"`
	EndAt        time.Time `db:"end_at"`
}
type bidDTO struct {
	ID        int   `db:"id"`
	AuctionID int   `db:"auction_id"`
	UserID    int   `db:"user_id"`
	Amount    int64 `db:"amount"`
}

type DB struct {
	log  *slog.Logger
	Conn *sqlx.DB
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

func (db DB) CreateAuction(ctx context.Context, auction core.Auction) (int, error) {
	query := `INSERT INTO auctions (sneaker_id, current_price, end_at) VALUES ($1, $2, $3) RETURNING id;`

	var id int

	err := db.Conn.QueryRowContext(ctx, query, auction.SneakerID, auction.CurrentPrice, auction.EndAt).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (db DB) GetAuctionByID(ctx context.Context, id int) (core.Auction, error) {
	query := `SELECT id, sneaker_id, current_price, end_at FROM auctions WHERE id=$1;`

	var dbRes auctionDTO

	err := db.Conn.GetContext(ctx, &dbRes, query, id)
	if err != nil {
		return core.Auction{}, err
	}

	res := core.Auction(dbRes)

	return res, nil
}

func (db DB) PlaceBid(ctx context.Context, bid core.Bid) error {
	tx, err := db.Conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		err := tx.Rollback()
		if err != nil {
			db.log.Error("error when rollback transaction", "error", err)
		}
	}()

	var cur_price int64
	var end_at time.Time

	blockQuery := `SELECT current_price, end_at FROM auctions WHERE id=$1 FOR UPDATE;`

	err = tx.QueryRowContext(ctx, blockQuery, bid.AuctionID).Scan(&cur_price, &end_at)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return core.ErrAuctionNotFound
		}
		return fmt.Errorf("error when select auction: %w", err)
	}

	if time.Now().After(end_at) {
		return core.ErrAuctionIsEnded
	}
	if cur_price > bid.Amount {
		return core.ErrBidToLow
	}

	updateQuery := `UPDATE auctions SET current_price=$1 WHERE id=$2;`

	_, err = tx.ExecContext(ctx, updateQuery, bid.Amount, bid.AuctionID)
	if err != nil {
		return fmt.Errorf("error when update auction: %w", err)
	}

	insertBidQuery := `INSERT INTO bids (auction_id, user_id, amount) VALUES ($1, $2, $3)`

	_, err = tx.ExecContext(ctx, insertBidQuery, bid.AuctionID, bid.UserID, bid.Amount)
	if err != nil {
		return fmt.Errorf("error when insert bid: %w", err)
	}

	return tx.Commit()
}

func (db DB) GetBidsByAuctionID(ctx context.Context, auctionID int) ([]core.Bid, error) {
	query := `SELECT id, auction_id, user_id, amount FROM bids WHERE auction_id=$1;`

	var dbBids []bidDTO

	err := db.Conn.SelectContext(ctx, &dbBids, query, auctionID)
	if err != nil {
		return []core.Bid{}, err
	}

	res := make([]core.Bid, 0, len(dbBids))

	for _, bid := range dbBids {
		res = append(res, core.Bid(bid))
	}

	return res, nil
}
