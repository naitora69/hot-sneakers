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

func (s Server) GetAllSneakers(ctx context.Context, in *catalogpb.GetAllSneakersRequest) (*catalogpb.GetAllSneakersResponse, error) {
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

	res := &catalogpb.GetAllSneakersResponse{
		Sneakers: sneakers,
	}

	return res, nil
}

func (s Server) GetSneakerByID(ctx context.Context, in *catalogpb.GetSneakerByIDRequest) (*catalogpb.GetSneakerByIDResponse, error) {
	sneakerFromService, err := s.service.GetSneakerByID(ctx, int(in.Id))
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	sneaker := &catalogpb.Sneaker{
		Id:    int64(sneakerFromService.ID),
		Brand: sneakerFromService.Brand,
		Model: sneakerFromService.Model,
	}

	return &catalogpb.GetSneakerByIDResponse{
		Sneaker: sneaker,
	}, nil
}

func (s Server) CreateSneaker(ctx context.Context, in *catalogpb.CreateSneakerRequest) (*catalogpb.CreateSneakerResponse, error) {
	sneaker := core.CreateSneaker{
		Brand: in.Brand,
		Model: in.Model,
	}

	id, err := s.service.CreateSneaker(ctx, sneaker)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &catalogpb.CreateSneakerResponse{
		Id: id,
	}, nil
}

func (s Server) UpdateSneaker(ctx context.Context, in *catalogpb.UpdateSneakerRequest) (*catalogpb.UpdateSneakerResponse, error) {
	resp := core.UpdateSneaker{
		ID:    in.Id,
		Brand: in.Brand,
		Model: in.Model,
	}

	err := s.service.UpdateSneaker(ctx, resp)
	if err != nil {
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &catalogpb.UpdateSneakerResponse{}, nil
}
