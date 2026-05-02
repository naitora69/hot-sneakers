package auction

import (
	"context"
	"log/slog"

	"hotsneakers/api/core"
	auctionpb "hotsneakers/proto/auction"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
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

func (c Client) GetAuctionDetails(ctx context.Context, auctionID int) (core.Auction, []core.Bid, error) {
	in := &auctionpb.GetAuctionDetailsReq{
		AuctionId: int64(auctionID),
	}

	resp, err := c.client.GetAuctionDetails(ctx, in)
	if err != nil {
		return core.Auction{}, nil, c.convertToCoreError(err)
	}

	endAt := resp.Auction.EndAt.AsTime()

	auction := core.Auction{
		ID:           int(resp.Auction.Id),
		SneakerID:    int(resp.Auction.SneakerId),
		CurrentPrice: resp.Auction.CurrentPrice,
		EndAt:        endAt,
	}

	bids := make([]core.Bid, 0, len(resp.Bids))
	for _, b := range resp.Bids {
		bids = append(bids, core.Bid{
			ID:     int(b.Id),
			UserID: int(b.UserId),
			Amount: b.Amount,
		})
	}

	return auction, bids, nil
}

func (c Client) MakeBid(ctx context.Context, auctionID int, userID int, amount int64) error {
	in := &auctionpb.MakeBidReq{
		AuctionId: int64(auctionID),
		UserId:    int64(userID),
		Amount:    amount,
	}

	_, err := c.client.MakeBid(ctx, in)
	if err != nil {
		return c.convertToCoreError(err)
	}

	return nil
}

func (c Client) CreateAuction(ctx context.Context, auction core.Auction) (int, error) {
	in := &auctionpb.CreateAuctionReq{
		SneakerId:    int64(auction.SneakerID),
		CurrentPrice: auction.CurrentPrice,
		EndAt:        timestamppb.New(auction.EndAt),
	}

	resp, err := c.client.CreateAuction(ctx, in)
	if err != nil {
		return 0, c.convertToCoreError(err)
	}

	return int(resp.AuctionId), nil
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
