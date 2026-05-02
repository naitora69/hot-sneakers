package catalog

import (
	"context"
	"log/slog"

	"hotsneakers/api/core"
	catalogpb "hotsneakers/proto/catalog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	log    *slog.Logger
	client catalogpb.CatalogClient
	Conn   *grpc.ClientConn
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := catalogpb.NewCatalogClient(conn)

	return &Client{
		log:    log,
		client: c,
		Conn:   conn,
	}, nil
}

func (c Client) Ping(ctx context.Context) error {
	_, err := c.client.Ping(ctx, nil)
	if err != nil {
		return c.convertToCoreError(err)
	}
	return nil
}

func (c Client) GetAllSneakers(ctx context.Context) ([]core.Sneaker, error) {
	resp, err := c.client.GetAllSneakers(ctx, &catalogpb.GetAllSneakersRequest{})
	if err != nil {
		return []core.Sneaker{}, c.convertToCoreError(err)
	}

	sneakers := make([]core.Sneaker, 0, len(resp.Sneakers))
	for _, sneaker := range resp.Sneakers {
		tmp := core.Sneaker{
			ID:    int(sneaker.Id),
			Brand: sneaker.Brand,
			Model: sneaker.Model,
		}
		sneakers = append(sneakers, tmp)
	}

	return sneakers, nil
}

func (c Client) GetSneakerByID(ctx context.Context, id int) (core.Sneaker, error) {
	in := &catalogpb.GetSneakerByIDRequest{
		Id: int64(id),
	}

	resp, err := c.client.GetSneakerByID(ctx, in)
	if err != nil {
		return core.Sneaker{}, c.convertToCoreError(err)
	}

	res := core.Sneaker{
		ID:    int(resp.Sneaker.Id),
		Brand: resp.Sneaker.Brand,
		Model: resp.Sneaker.Model,
	}

	return res, nil
}

func (c Client) CreateSneaker(ctx context.Context, sneaker core.CreateSneaker) (int64, error) {
	in := &catalogpb.CreateSneakerRequest{
		Brand: sneaker.Brand,
		Model: sneaker.Model,
	}

	resp, err := c.client.CreateSneaker(ctx, in)
	if err != nil {
		return 0, c.convertToCoreError(err)
	}

	return resp.Id, nil
}

func (c Client) UpdateSneaker(ctx context.Context, sneaker core.UpdateSneaker) error {
	in := &catalogpb.UpdateSneakerRequest{
		Id:    sneaker.ID,
		Brand: sneaker.Brand,
		Model: sneaker.Model,
	}

	_, err := c.client.UpdateSneaker(ctx, in)
	if err != nil {
		return err
	}

	return nil
}

func (c Client) convertToCoreError(err error) error {
	s, ok := status.FromError(err)
	if !ok {
		return core.ErrInternal
	}

	switch s.Code() {
	case codes.Canceled:
		return core.ErrCanceled
	case codes.DeadlineExceeded:
		return core.ErrDeadlineExceeded
	default:
		c.log.Error("unexpected grpc error", "code", s.Code(), "msg", s.Message())
		return core.ErrInternal
	}
}
