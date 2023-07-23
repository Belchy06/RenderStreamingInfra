package main

import (
	"log"

	"github.com/pion/webrtc/v3"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"tensorworks.com.au/cumulus/signalling"
)

type Peer struct {
	pc     *webrtc.PeerConnection
	err    error
	sdp    string
	client signalling.SignallingClient
	msg    chan string
}

func (p *Peer) listenForOffers() {
	ctx := context.Background()
	offerStream, err := p.client.SubscribeToApplicationOffer(ctx, &signalling.Empty{})
	if err != nil {
		log.Fatal(err)
	}

	for {
		log.Println("listening for offers")
		offer, err := offerStream.Recv()
		if err != nil {
			log.Fatal(err)
		}

		log.Println("received offer")
		log.Println(offer.Sdp)
		break
	}
}

func main() {
	done := make(chan bool, 0)

	dialOptions := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithUserAgent("grpc-go-client"),
	}
	ctx := context.Background()

	cc, err := grpc.DialContext(ctx, ":8080", dialOptions...)
	if err != nil {
		log.Fatal(err)
	}

	client := signalling.NewSignallingClient(cc)

	signallingConfig, err := client.Config(ctx, &signalling.Empty{})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(signallingConfig.PeerConnectionOptions)

	// TODO (belchy06): Parse `signallingConfig.PeerConnectionOptions` for peerConnectionOptions
	config := webrtc.Configuration{ICEServers: []webrtc.ICEServer{{URLs: []string{"stun:stun.l.google.com:19302"}}}}

	pc, err := webrtc.NewPeerConnection(config)
	if nil != err {
		log.Println("Failed to create PeerConnection.")
		return
	}

	peer := &Peer{
		client: client,
		pc:     pc,
	}

	// TODO (belchy06): Setup all the subscriptions before connecting
	go peer.listenForOffers()

	// Connect!
	_, err = client.ConnectPlayer(ctx, &signalling.Empty{})
	if err != nil {
		log.Fatal(err)
	}

	<-done
}
