package grpc

import (
	"context"
	"hotsneakers/catalog/core"
	catalogpb "hotsneakers/proto/catalog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	catalogpb.UnimplementedCatalogServer
	service core.Cataloger
}

func NewServer(service core.Cataloger) *Server {
	return &Server{
		service: service,
	}
}

func (s Server) Ping(_ context.Context, _ *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, nil
}

func (s Server) GetAllSneakers(ctx context.Context, in *catalogpb.GetAllSneakersRequest) (*catalogpb.GetAllSneakersReply, error) {
	sneakersFromService, err := s.service.GetAllSneakers(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	sneakers := make([]*catalogpb.Sneaker, 0, len(sneakersFromService))

	for _, sneaker := range sneakersFromService {
		sneakers = append(sneakers, &catalogpb.Sneaker{
			Id:    int64(sneaker.ID),
			Brand: sneaker.Brand,
			Model: sneaker.Model,
		})
	}

	res := &catalogpb.GetAllSneakersReply{
		Sneakers: sneakers,
	}

	return res, nil
}
func (s Server) GetSneakerByID(ctx context.Context, in *catalogpb.GetSneakerByIDRequest) (*catalogpb.Sneaker, error) {
	sneakerFromService, err := s.service.GetSneakerByID(ctx, int(in.Id))
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	res := &catalogpb.Sneaker{
		Id:    int64(sneakerFromService.ID),
		Brand: sneakerFromService.Brand,
		Model: sneakerFromService.Model,
	}

	return res, nil
}
