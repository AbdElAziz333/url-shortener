package server

import (
	"context"
	"errors"
	"fmt"
	"net"

	"aziz.dev/shortener/internal/link"
	"aziz.dev/shortener/internal/pb"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type grpcServer struct {
	pb.UnimplementedShortenerServiceServer
	service link.Service
}

func NewGRPCServer(service link.Service) *grpcServer {
	return &grpcServer{service: service}
}

func StartGRPC(service link.Service, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("listen on port %s: %w", port, err)
	}

	srv := grpc.NewServer()
	pb.RegisterShortenerServiceServer(srv, NewGRPCServer(service))

	logrus.WithField("port", port).Info("gRPC server listening")
	return srv.Serve(lis)
}

func (s *grpcServer) GetAllLinks(ctx context.Context, req *pb.GetAllLinksRequest) (*pb.GetAllLinksResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	limit := req.GetLimit()

	if limit <= 0 {
		limit = 20
	}

	offset := req.GetOffset()
	links, err := s.service.GetAll(ctx, userID, limit, offset)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp := &pb.GetAllLinksResponse{Links: make([]*pb.Link, 0, len(links))}
	for i := range links {
		resp.Links = append(resp.Links, dtoToProto(&links[i]))
	}

	return resp, nil
}

func (s *grpcServer) GetLinkByCode(ctx context.Context, req *pb.GetLinkByCodeRequest) (*pb.GetLinkByCodeResponse, error) {
	uid, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
 
	dto, err := s.service.Get(ctx, uid, req.GetCode())
	if err != nil {
		return nil, toStatus(err)
	}
	return &pb.GetLinkByCodeResponse{Link: dtoToProto(dto)}, nil
}

func (s *grpcServer) CreateLink(ctx context.Context, req *pb.CreateLinkRequest) (*pb.CreateLinkResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	createReq := link.CreateRequest{
		OriginalURL: req.GetOriginalUrl(),
	}
	if req.ExpiresAt != nil {
		t := req.ExpiresAt.AsTime()
		createReq.ExpiresAt = &t
	}

	created, err := s.service.Create(ctx, userID, createReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.CreateLinkResponse{Link: dtoToProto(created)}, nil
}

func (s *grpcServer) UpdateExpiry(ctx context.Context, req *pb.UpdateExpiryRequest) (*pb.UpdateExpiryResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	if req.ExpiresAt == nil {
		return nil, status.Error(codes.InvalidArgument, "expires_at is required")
	}

	expiresAt := req.ExpiresAt.AsTime()
	updateReq := link.UpdateExpiryRequest{ExpiresAt: &expiresAt}

	_, err = s.service.UpdateExpiry(ctx, userID, req.GetCode(), updateReq)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.UpdateExpiryResponse{Message: "successfully updated expiry"}, nil
}

func (s *grpcServer) DeleteLink(ctx context.Context, req *pb.DeleteLinkRequest) (*pb.DeleteLinkResponse, error) {
	userID, err := uuid.Parse(req.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid user id")
	}

	if err := s.service.Delete(ctx, userID, req.GetCode()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.DeleteLinkResponse{Message: "successfully deleted link"}, nil
}

func dtoToProto(d *link.Dto) *pb.Link {
	return &pb.Link{
		Code:        d.Code,
		OriginalUrl: d.OriginalURL,
		ExpiresAt: timestamppb.New(d.ExpiresAt),
		IsActive:    d.IsActive,
		CreatedAt:   timestamppb.New(d.CreatedAt),
	}
}

func parseUserID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	return id, nil
}
 
// toStatus maps domain errors to gRPC codes; unexpected errors are logged once here.
func toStatus(err error) error {
	switch {
	case errors.Is(err, link.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, link.ErrInvalidExpiry):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		logrus.WithError(err).Error("request failed")
		return status.Error(codes.Internal, "internal server error")
	}
}