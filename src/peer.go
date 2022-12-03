package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	filepath "path/filepath"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
	dutil "github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"
)

// Mock Data
var (
	BootstrapIP = "127.0.0.1"

	//IDS do PedroS, é preciso mudar para os ids dos nossos nodes de bootstrap ou por um mecânismo que leia os ids dos nodes de bootstrap apartir dos files
	BootstrapNodes = []BoostrapNode{}
)

func peerNodeInit(register bool, username string, password string, port int) {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	getBootstrapIds()

	var idFilePath = "./nodes/" + username + ".json"
	if register {
		//TODO

		// if _, err := os.Stat(idFilePath); err == nil {
		// 	// Generate a new key pair for this host. We will use it to obtain a valid host ID.
		// 	r := rand.New(rand.NewSource(time.Now().UnixNano()))
		// 	prvKey, pubKey, err = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)

		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	//check if username already in use
		// 	host, err := makeHost(port, prvKey)

		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	_, err = dht.New(ctx, host)
		// 	if err != nil {
		// 		panic(err)
		// 	}

		// 	//send message to dht with the username and check if it is already in use

		// 	//if username is not in use, save the keypair and create a new account

		// 	if saveNodeState(idFilePath, prvKey, pubKey) != nil {
		// 		panic(err)
		// 	}

		// 	//else, return error

		// 	panic(err)
		// } else {
		// 	panic(err)

		// }

	} else {

		prvKey, _, err := loadNodeState(idFilePath)
		if err != nil {
			panic(err)
		}
		host, err := makeHost(port, prvKey)
		if err != nil {
			panic(err)
		}

		var kademliaDHT *dht.IpfsDHT = initDHT(ctx, host)

		channel := make(chan string)
		go discoverPeers(channel, ctx, host, kademliaDHT)
		select {
		case msg1 := <-channel:
			fmt.Println(msg1)
		case <-time.After(3 * time.Second):
			fmt.Println("Timeout!")
		}

		ps, err := pubsub.NewGossipSub(ctx, host)
		if err != nil {
			panic(err)
		}
		//Joins own topic
		fmt.Println("here4")

		topic_own, err := ps.Join(username)
		if err != nil {
			panic(err)
		}
		fmt.Println("here5")

		topic_general, err := ps.Join("rettiwt")
		if err != nil {
			panic(err)
		}

		fmt.Println("here6")

		_, err = topic_own.Subscribe()
		if err != nil {
			panic(err)
		}

		_, err = topic_general.Subscribe()
		if err != nil {
			panic(err)
		}
	}

	select {}

}

func getBootstrapIds() {

	nodes, err := os.ReadFile("./nodes/bootstrap_nodes.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(nodes, &BootstrapNodes)
	if err != nil {
		panic(err)
	}

}

func initDHT(ctx context.Context, h host.Host) *dht.IpfsDHT {
	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	kademliaDHT, err := dht.New(ctx, h)
	fmt.Println("here1")
	if err != nil {
		panic(err)
	}
	fmt.Println("here2")

	if err = kademliaDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}
	fmt.Println("here3")

	var wg sync.WaitGroup
	for _, bsnode := range BootstrapNodes {
		BSMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/%s/tcp/700%d/p2p/%s", BootstrapIP, bsnode.Port, bsnode.Id))
		peerinfo, _ := peer.AddrInfoFromP2pAddr(BSMultiAddr)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := h.Connect(ctx, *peerinfo); err != nil {
				logger.Warning(err)
			} else {
				logger.Info("Connection established with bootstrap node:", *peerinfo)
			}
		}()
	}
	wg.Wait()

	return kademliaDHT
}

func discoverPeers(c chan string, ctx context.Context, h host.Host, kademliaDHT *dht.IpfsDHT) {
	routingDiscovery := drouting.NewRoutingDiscovery(kademliaDHT)
	dutil.Advertise(ctx, routingDiscovery, "rettiwt")

	fmt.Printf("Successfully announced ourselves as a bootstrap node\n")

	// Look for others who have announced and attempt to connect to them
	anyConnected := false
	for !anyConnected {
		fmt.Println("Searching for peers...")
		peerChan, err := routingDiscovery.FindPeers(ctx, "rettiwt")
		if err != nil {
			panic(err)
		}
		for peer := range peerChan {
			if peer.ID == h.ID() {
				continue // No self connection
			}
			err := h.Connect(ctx, peer)
			if err != nil {
				fmt.Println("Failed connecting to ", peer.ID.Pretty(), ", error:", err)
			} else {
				fmt.Println("Connected to:", peer.ID.Pretty())
				anyConnected = true
			}
		}
	}
	c <- "Peer discovery complete"
}
func makeHost(port int, prvKey crypto.PrivKey) (host.Host, error) {
	// Creates a new RSA key pair for this host.

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	return libp2p.New(
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)
}

func loadNodeState(idFilePath string) (crypto.PrivKey, crypto.PubKey, error) {
	var pair KeyPair
	jsonPair, err := os.ReadFile(idFilePath)
	if err != nil {
		return nil, nil, err
	}

	err = json.Unmarshal(jsonPair, &pair)
	if err != nil {
		return nil, nil, err
	}

	prvKey, err := crypto.UnmarshalPrivateKey(pair.PrvKey)
	if err != nil {
		return nil, nil, err
	}

	pubKey, err := crypto.UnmarshalPublicKey(pair.PubKey)
	if err != nil {
		return nil, nil, err
	}
	return prvKey, pubKey, nil
}

func saveNodeState(idFilePath string, prvKey crypto.PrivKey, pubKey crypto.PubKey) error {
	prvKeyBytes, err := crypto.MarshalPrivateKey(prvKey)
	if err != nil {
		return err
	}
	pubKeyBytes, err := crypto.MarshalPublicKey(pubKey)
	if err != nil {
		return err
	}

	pair := KeyPair{
		PrvKey: prvKeyBytes,
		PubKey: pubKeyBytes,
	}

	json_pair, err := json.MarshalIndent(pair, "", "    ")
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(idFilePath), 0777)
	if err != nil {
		return err
	}

	err = os.WriteFile(idFilePath, json_pair, 0666)
	if err != nil {
		return err
	}
	return nil
}
