package grpc

import (
	"context"

	"hotsneakers/auction/core"
	auctionpb "hotsneakers/proto/auction"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	auctionpb.UnimplementedAuctionServiceServer
	service core.AuctionService
}

func NewServer(service core.AuctionService) *Server {
	return &Server{
		service: service,
	}
}

func (s Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (s Server) CreateAuction(ctx context.Context, in *auctionpb.CreateAuctionReq) (*auctionpb.CreateAuctionResp, error) {
	auctionCore := core.Auction{
		SneakerID:    int(in.SneakerId),
		CurrentPrice: in.CurrentPrice,
		EndAt:        in.EndAt.AsTime(),
	}

	auctionID, err := s.service.CreateAuction(ctx, auctionCore)
	if err != nil {
		return nil, convertCoreErrorToGrpc(err)
	}

	out := &auctionpb.CreateAuctionResp{
		AuctionId: int64(auctionID),
	}

	return out, nil
}

func (s Server) GetAuctionDetails(ctx context.Context, in *auctionpb.GetAuctionDetailsReq) (*auctionpb.GetAuctionDetailsResp, error) {
	auctionCore, bidsCore, err := s.service.GetAuctionDetails(ctx, int(in.AuctionId))
	if err != nil {
		return nil, convertCoreErrorToGrpc(err)
	}

	pbAuction := &auctionpb.Auction{
		Id:           int64(auctionCore.ID),
		SneakerId:    int64(auctionCore.SneakerID),
		CurrentPrice: auctionCore.CurrentPrice,
		EndAt:        timestamppb.New(auctionCore.EndAt),
	}

	pdBids := make([]*auctionpb.Bid, 0, len(bidsCore))

	for _, b := range bidsCore {
		pdBids = append(pdBids, &auctionpb.Bid{
			Id:        int64(b.ID),
			UserId:    int64(b.UserID),
			AuctionId: int64(b.AuctionID),
			Amount:    b.Amount,
		})
	}

	out := &auctionpb.GetAuctionDetailsResp{
		Auction: pbAuction,
		Bids:    pdBids,
	}

	return out, nil
}

func (s Server) MakeBid(ctx context.Context, in *auctionpb.MakeBidReq) (*auctionpb.MakeBidResp, error) {
	if in.Amount <= 0 {
		return nil, status.Error(codes.InvalidArgument, "amount must be greater than 0")
	}

	err := s.service.MakeBid(ctx, int(in.AuctionId), int(in.UserId), in.Amount)
	if err != nil {
		return nil, convertCoreErrorToGrpc(err)
	}

	return &auctionpb.MakeBidResp{}, nil
}

func convertCoreErrorToGrpc(err error) error {
	switch err {
	case core.ErrAuctionIsEnded:
		return status.Error(codes.FailedPrecondition, "auction has already ended")
	case core.ErrAuctionNotFound:
		return status.Error(codes.NotFound, "error auction not found")
	case core.ErrBidToLow:
		return status.Error(codes.FailedPrecondition, "bid amount is too low")
	case core.ErrInvalidAuctionOrUserId:
		return status.Error(codes.InvalidArgument, "invalid auction or user id")
	}

	return status.Error(codes.Internal, err.Error())
}
