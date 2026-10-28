package grpcclient

import (
	"errors"
	"fmt"

	linkpb "aziz.dev/gateway/internal/pb/link"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Link linkpb.LinkServiceClient
	conns     []*grpc.ClientConn
}

func NewClients(shortenerAddr string) (*Clients, error) {
	shortenerConn, err := dial(shortenerAddr)
	if err != nil {
		return nil, fmt.Errorf("dial shortener: %w", err)
	}

	return &Clients{
		Link: linkpb.NewLinkServiceClient(shortenerConn),
		conns:     []*grpc.ClientConn{shortenerConn},
	}, nil
}

func (c *Clients) Close() error {
	var errs []error
	for _, conn := range c.conns {
		errs = append(errs, conn.Close())
	}
	return errors.Join(errs...)
}

func dial(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}