package peer

import (
	"context"
	"errors"
	"fmt"

	log "github.com/ipfs/go-log/v2"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/auth"
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht"
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht/record/account"
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/peer"
)

var peerLogger = log.Logger("rettiwt-peer")
var ErrCannotLogin = errors.New("login: cannot login: ")

func NodeInit(ctx context.Context, idFilePath, bootstrapPeerIdsFilePath string, port int) (host.Host, *dht.KademliaDHT) {

	// Create the host
	host := peer.CreateHost(ctx, idFilePath, port)

	var dht = peer.InitDHT("rettiwt", ctx, host, bootstrapPeerIdsFilePath)

	go peer.DiscoverPeers(ctx, host, dht, "rettiwt")

	fmt.Println("Idle...")

	return host, dht
}

func RegisterUser(register bool, dht *dht.KademliaDHT, username, password string) error {
	// Check if user is already registered
	exists, err := dht.KeyExists(account.AccountNS + username)
	if err != nil {
		peerLogger.Error(err)
		return errors.New("register: cannot check username existence: " + err.Error())
	}

	if !exists {
		if !register {
			peerLogger.Panic(errors.New("Username doesn't exist and no register requested: " + auth.ErrUsernameDoesntExist.Error()))
		}

		// Register user
		err := auth.Register(dht, username, password)
		if err != nil {
			err = errors.New("register: cannot register: " + err.Error())
			peerLogger.Error(err)
			return err
		}
	}

	return nil
}

func LoginUser(dht *dht.KademliaDHT, username, password string) error {
	// Login user
	err := auth.Login(dht, username, password)
	if err != nil {
		err = errors.New("login: cannot login: " + err.Error())
		peerLogger.Error(err)
		return err
	}

	return nil
}

func PubSubInit(ctx context.Context, host host.Host, username, idFilePath string) *pubsub.PubSub {
	tracer, err := pubsub.NewJSONTracer(idFilePath + ".trace.json")
	if err != nil {
		peerLogger.Panic(errors.New("Can't create JSON tracer: " + err.Error()))
	}

	pubSub, err := pubsub.NewGossipSub(ctx, host, pubsub.WithEventTracer(tracer))
	if err != nil {
		peerLogger.Panic(errors.New("Can't create GossipPubSub: " + err.Error()))
	}

	return pubSub
}
