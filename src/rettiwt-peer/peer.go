package peer

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	pubsub "github.com/libp2p/go-libp2p-pubsub"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/peer"
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/timeline"
)

// var logger = log.Logger("rettiwt-peer")

func NodeInit(idFilePath, bootstrapPeerIdsFilePath string, register bool, username string, password string, port int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the host
	host := peer.CreateHost(ctx, idFilePath, port)

	var dht = peer.InitDHT("rettiwt", ctx, host, bootstrapPeerIdsFilePath)

	go peer.DiscoverPeers(ctx, host, dht, "rettiwt")

	fmt.Println("Connected to bootstrap nodes")

	// channel := make(chan string)
	// go discoverPeers(channel, ctx, host, kademliaDHT)
	// select {
	// case msg1 := <-channel:
	// 	fmt.Println(msg1)
	// case <-time.After(3 * time.Second):
	// 	fmt.Println("Timeout!")
	// }

	ps, err := pubsub.NewGossipSub(ctx, host)
	if err != nil {
		panic(err)
	}
	//Joins own topic
	timeline.FollowUser(ctx, ps, host.ID(), username)
	timeline.FollowUser(ctx, ps, host.ID(), "rettitw")

	// if register {
	// 	//TODO

	// 	// if _, err := os.Stat(idFilePath); err == nil {
	// 	// 	// Generate a new key pair for this host. We will use it to obtain a valid host ID.
	// 	// 	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 	// 	prvKey, pubKey, err = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)

	// 	// 	if err != nil {
	// 	// 		panic(err)
	// 	// 	}

	// 	// 	//check if username already in use
	// 	// 	host, err := makeHost(port, prvKey)

	// 	// 	if err != nil {
	// 	// 		panic(err)
	// 	// 	}

	// 	// 	_, err = dht.New(ctx, host)
	// 	// 	if err != nil {
	// 	// 		panic(err)
	// 	// 	}

	// 	// 	//send message to dht with the username and check if it is already in use

	// 	// 	//if username is not in use, save the keypair and create a new account

	// 	// 	if saveNodeState(idFilePath, prvKey, pubKey) != nil {
	// 	// 		panic(err)
	// 	// 	}

	// 	// 	//else, return error

	// 	// 	panic(err)
	// 	// } else {
	// 	// 	panic(err)

	// 	// }

	// } else {

	// }

	fmt.Println("Idle...")

	var text string

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Please enter command (help | publish <string> | follow <string> | unfollow <string> | update) : ")
		text, _ = reader.ReadString('\n')
		text = strings.Replace(text, "\n", "", -1)
		words := strings.Fields(text)

		switch words[0] {
		case "publish":
			err := timeline.Publish(words[1], username)
			if err != nil {
				fmt.Println(err)
			}
		case "follow":
			timeline.FollowUser(ctx, ps, host.ID(), words[1])
		case "unfollow":
			timeline.UnfollowUser(ctx, ps, words[1])
		case "update":
			timeline.UpdateTimeline()
		case "help":
			fmt.Println("publish <string> - Publishes a tweet")
			fmt.Println("follow <string> - Follows a user")
			fmt.Println("unfollow <string> - Unfollows a user")
			fmt.Println("update - Updates the timeline")
		default:
			fmt.Println("Invalid command")
		}
	}

}
