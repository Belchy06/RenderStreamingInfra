package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"tensorworks.com.au/cumulus/signalling"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

type SignallingServer struct {
	signalling.SignallingServer
	currentPlayerId int
	streamers       map[string]Streamer
	players         map[string]Player
}

type Player struct {
	applicationOffer chan string
}

type Streamer struct {
	playerConnected chan string
}

func (s *SignallingServer) Config(ctx context.Context, in *signalling.Empty) (*signalling.PeerConfig, error) {
	return &signalling.PeerConfig{PeerConnectionOptions: "{ \"iceServers\": [{\"urls\": [\"stun: stun.l.google.com:19302\"]}]}"}, nil
}

/*
Player
*/
func (s *SignallingServer) ConnectPlayer(ctx context.Context, in *signalling.Empty) (*signalling.Empty, error) {
	log.Printf("player \"%d\" connected\n", s.currentPlayerId)

	s.players[fmt.Sprintf("%d", s.currentPlayerId)] = Player{
		applicationOffer: make(chan string, 0),
	}

	for _, streamer := range s.streamers {
		streamer.playerConnected <- fmt.Sprintf("%d", s.currentPlayerId)
	}

	s.currentPlayerId = s.currentPlayerId + 1
	return &signalling.Empty{}, nil
}

func (s *SignallingServer) SubscribeToApplicationOffer(empty *signalling.Empty, stream signalling.Signalling_SubscribeToApplicationOfferServer) error {
	for {
		for id, player := range s.players {
			select {
			case offer := <-player.applicationOffer:
				log.Printf("cumulus -> %s: PlayerConnected\n", id)
				err := stream.Send(&signalling.Offer{
					Sdp: offer,
				})
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

/*
Application
*/
func (s *SignallingServer) ConnectStreamer(ctx context.Context, in *signalling.Streamer) (*signalling.Empty, error) {
	log.Printf("streamer \"%s\" connected\n", in.Id)
	s.streamers[in.Id] = Streamer{
		playerConnected: make(chan string, 0),
	}
	return &signalling.Empty{}, nil
}

func (s *SignallingServer) SubscribeToPlayerConnected(empty *signalling.Empty, stream signalling.Signalling_SubscribeToPlayerConnectedServer) error {
	for {
		for _, streamer := range s.streamers {
			select {
			case playerId := <-streamer.playerConnected:
				log.Printf("cumulus -> streamer: PlayerConnected\n")
				err := stream.Send(&signalling.PlayerConnected{
					Id: playerId,
				})
				if err != nil {
					log.Fatal(err)
				}
			}
		}
	}
}

func (s *SignallingServer) SendOfferToPlayer(ctx context.Context, in *signalling.Offer) (*signalling.Empty, error) {
	log.Printf("streamer -> cumulus: SendOfferToPlayer\n")
	player := s.players[in.Id]
	player.applicationOffer <- in.Sdp

	return &signalling.Empty{}, nil
}

func main() {
	listener, err := net.Listen("tcp", ":8080")

	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()

	ss := &SignallingServer{
		currentPlayerId: 100,
		players:         make(map[string]Player),
		streamers:       make(map[string]Streamer),
	}

	signalling.RegisterSignallingServer(s, ss)

	if err := s.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
