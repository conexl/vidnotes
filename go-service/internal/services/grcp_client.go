package services

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClientService interface {
	GetConnection() *grpc.ClientConn
	Close() error
}

type grpcClientService struct {
	conn *grpc.ClientConn
}

func NewGRPCClientService(addr string) (GRPCClientService, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(500*1024*1024),
			grpc.MaxCallSendMsgSize(500*1024*1024),
		),
	)
	if err != nil {
		return nil, err
	}

	return &grpcClientService{
		conn: conn,
	}, nil
}

func (g *grpcClientService) GetConnection() *grpc.ClientConn {
	return g.conn
}

func (g *grpcClientService) Close() error {
	return g.conn.Close()
}
