package peer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	log "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
	"github.com/multiformats/go-multiaddr"

	coreDHT "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht"
)

var creationLogger = log.Logger("creation")

func CreateHost(ctx context.Context, idFilePath string, port int) host.Host {
	var prvKey crypto.PrivKey
	if _, err := os.Stat(idFilePath); err != nil {
		prvKey, _ = CreateIdentity(idFilePath)
	} else {
		prvKey, _ = LoadIdentity(idFilePath)
	}

	var networkNotifiee network.NotifyBundle
	networkNotifiee.ListenF = func(net network.Network, ma multiaddr.Multiaddr) {
		creationLogger.Info("Listening on %s, on interface %s", ma, net)
	}

	networkNotifiee.ConnectedF = func(net network.Network, con network.Conn) {
		creationLogger.Info("Connected to %s on interface %s", con, net)
	}

	// 0.0.0.0 will listen on any interface device.
	hostMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	host, err := libp2p.New(libp2p.ListenAddrs(hostMultiAddr), libp2p.Identity(prvKey))
	if err != nil {
		creationLogger.Panic(err)
	}

	host.Network().Notify(&networkNotifiee)

	return host
}

func InitDHT(ctx context.Context, host host.Host, bootstrapPeerIdsFilePath string) *coreDHT.KademliaDHT {
	bootstrapPeerIds, err := getBootstrapNodesList(bootstrapPeerIdsFilePath)
	if err != nil {
		creationLogger.Error("Could not open node multiaddrs list", err)
	}
	var bootstrapPeers []peer.AddrInfo
	for _, bootstrapPeer := range bootstrapPeerIds {
		bootstrapNodeMultiAddr, _ := multiaddr.NewMultiaddr(bootstrapPeer)
		peerinfo, err := peer.AddrInfoFromP2pAddr(bootstrapNodeMultiAddr)
		if err != nil {
			creationLogger.Error("Error parsing bootstrap peer address", err)
			continue
		}

		bootstrapPeers = append(bootstrapPeers, *peerinfo)
	}

	// Start a DHT, for use in peer discovery. We can't just make a new DHT
	// client because we want each peer to maintain its own local copy of the
	// DHT, so that the bootstrapping node of the DHT can go down without
	// inhibiting future peer discovery.
	rettiwtDHT, err := coreDHT.NewKademliaDHT(host, ctx, dht.BootstrapPeers(bootstrapPeers...))
	if err != nil {
		creationLogger.Panic(err)
	}

	if err = rettiwtDHT.IpfsDHT.Bootstrap(ctx); err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	for _, bootstrapPeer := range bootstrapPeers {
		bootPeerRef := bootstrapPeer
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := host.Connect(ctx, bootPeerRef); err != nil {
				creationLogger.Warning("Error connecting to bootstrap node", err)
			} else {
				creationLogger.Info("Connection established with bootstrap node:", bootPeerRef)
			}
		}()
	}
	wg.Wait()

	return rettiwtDHT
}

func getBootstrapNodesList(bootstrapPeerIdsFilePath string) ([]string, error) {
	var bootstrapNodes []string
	nodes, err := os.ReadFile(bootstrapPeerIdsFilePath)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(nodes, &bootstrapNodes)
	if err != nil {
		return nil, err
	}
	return bootstrapNodes, nil
}

func DiscoverPeers(ctx context.Context, h host.Host, dht *coreDHT.KademliaDHT) {
	routingDiscovery := routing.NewRoutingDiscovery(dht.IpfsDHT)
	util.Advertise(ctx, routingDiscovery, "rettiwt")

	// Look for others who have announced and attempt to connect to them
	anyConnected := false
	for !anyConnected {
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
				creationLogger.Error("Failed connecting to ", peer.ID.Pretty(), ", error:", err)
			} else {
				creationLogger.Info("Connected to:", peer.ID.Pretty())
				anyConnected = true
			}
		}
	}
	fmt.Println("Peer discovery complete")
}
