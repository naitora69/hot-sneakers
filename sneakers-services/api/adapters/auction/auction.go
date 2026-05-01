package auction

import (
	"context"
	"hotsneakers/api/core"
	auctionpb "hotsneakers/proto/auction"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Client struct {
	log    *slog.Logger
	client auctionpb.AuctionServiceClient
	Conn   *grpc.ClientConn
}

func NewClient(address string, log *slog.Logger) (*Client, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	c := auctionpb.NewAuctionServiceClient(conn)

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
