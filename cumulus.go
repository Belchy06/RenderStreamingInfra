package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"tensorworks.com.au/cumulus/signalling"
)

type server struct {
	signalling.SignallingServiceServer
}

func (s *server) Ping(ctx context.Context, in *signalling.PingMsg) (*signalling.PongMsg, error) {
	log.Printf("Ping message body from client: %s", in.Message)
	return &signalling.PongMsg{Message: "Pong"}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()

	signalling.RegisterSignallingServiceServer(grpcServer, &server{})

	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
