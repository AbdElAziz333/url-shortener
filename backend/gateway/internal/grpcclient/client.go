package grpcclient

import (
	"errors"
	"fmt"

	shortenerpb "aziz.dev/gateway/internal/pb/shortener"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Clients struct {
	Shortener shortenerpb.ShortenerServiceClient
	conns     []*grpc.ClientConn
}

func NewClients(shortenerAddr string) (*Clients, error) {
	shortenerConn, err := dial(shortenerAddr)
	if err != nil {
		return nil, fmt.Errorf("dial shortener: %w", err)
	}

	return &Clients{
		Shortener: shortenerpb.NewShortenerServiceClient(shortenerConn),
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