package link

import (
	"context"
	"errors"
	"fmt"
	"net"

	"aziz.dev/shortener/internal/pb"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	defaultLimit = 20
	maxLimit     = 100
)

type grpcServer struct {
	pb.UnimplementedLinkServiceServer
	service Service
}

func NewGRPCServer(service Service) *grpcServer {
	return &grpcServer{service: service}
}

func StartGRPC(service Service, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("listen on port %s: %w", port, err)
	}

	srv := grpc.NewServer()
	pb.RegisterLinkServiceServer(srv, NewGRPCServer(service))

	logrus.WithField("port", port).Info("gRPC server listening")
	return srv.Serve(lis)
}

func (s *grpcServer) ListLinks(ctx context.Context, req *pb.ListLinksRequest) (*pb.ListLinksResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}

	limit := req.GetLimit()
	if limit <= 0 {
		limit = defaultLimit
	} else if limit > maxLimit {
		limit = maxLimit
	}

	offset := req.GetOffset()
	if offset < 0 {
		offset = 0
	}

	links, err := s.service.GetAll(ctx, userID, limit, offset)
	if err != nil {
		return nil, toStatus(err, userID)
	}

	resp := &pb.ListLinksResponse{Links: make([]*pb.Link, 0, len(links))}
	for i := range links {
		resp.Links = append(resp.Links, toProto(&links[i]))
	}
	return resp, nil
}

func (s *grpcServer) GetLink(ctx context.Context, req *pb.GetLinkRequest) (*pb.GetLinkResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}

	l, err := s.service.Get(ctx, userID, req.GetCode())
	if err != nil {
		return nil, toStatus(err, userID)
	}
	return &pb.GetLinkResponse{Link: toProto(l)}, nil
}

func (s *grpcServer) CreateLink(ctx context.Context, req *pb.CreateLinkRequest) (*pb.CreateLinkResponse, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}
	if req.GetExpiresAt() == nil {
		return nil, status.Error(codes.InvalidArgument, "expires_at is required")
	}

	created, err := s.service.Create(ctx, userID, req.GetOriginalUrl(), req.GetExpiresAt().AsTime())
	if err != nil {
		return nil, toStatus(err, userID)
	}
	return &pb.CreateLinkResponse{Link: toProto(created)}, nil
}

func (s *grpcServer) UpdateLinkExpiry(ctx context.Context, req *pb.UpdateLinkExpiryRequest) (*emptypb.Empty, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}

	if req.ExpiresAt == nil {
		return nil, status.Error(codes.InvalidArgument, "expires_at is required")
	}

	if _, err := s.service.UpdateExpiry(ctx, userID, req.GetCode(), req.GetExpiresAt().AsTime()); err != nil {
		return nil, toStatus(err, userID)
	}

	return &emptypb.Empty{}, nil
}

func (s *grpcServer) DeleteLink(ctx context.Context, req *pb.DeleteLinkRequest) (*emptypb.Empty, error) {
	userID, err := parseUserID(req.GetUserId())
	if err != nil {
		return nil, err
	}

	if err := s.service.Delete(ctx, userID, req.GetCode()); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &emptypb.Empty{}, nil
}

// toProto is the domain -> wire boundary.
func toProto(l *Link) *pb.Link {
	out := &pb.Link{
		Code:        l.Code,
		OriginalUrl: l.OriginalURL,
		IsActive:    l.IsActive,
		CreatedAt:   timestamppb.New(l.CreatedAt),
	}
	// leave expires_at unset when there's no expiry,
	// otherwise a zero time becomes year 1
	if !l.ExpiresAt.IsZero() {
		out.ExpiresAt = timestamppb.New(l.ExpiresAt)
	}
	return out
}

func parseUserID(s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, status.Error(codes.InvalidArgument, "invalid user id")
	}
	return id, nil
}
 
// toStatus maps domain errors to gRPC codes; unexpected errors are logged once
// here and returned with a generic message so internals don't leak.
func toStatus(err error, userID uuid.UUID) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, ErrInvalidExpiry):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		logrus.WithError(err).WithField("user_id", userID).Error("request failed")
		return status.Error(codes.Internal, "internal server error")
	}
}