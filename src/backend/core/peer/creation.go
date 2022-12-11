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
	"github.com/libp2p/go-libp2p/core/peer"
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

	// 0.0.0.0 will listen on any interface device.
	hostMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	host, err := libp2p.New(libp2p.ListenAddrs(hostMultiAddr), libp2p.Identity(prvKey))
	if err != nil {
		creationLogger.Panic(err)
	}

	networkNotifiee := GetNotifiee()
	host.Network().Notify(&networkNotifiee)

	return host
}

func InitDHT(mode string, ctx context.Context, host host.Host, bootstrapPeerIdsFilePath string) *coreDHT.KademliaDHT {
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

		if peerinfo.ID == host.ID() {
			continue
		}

		bootstrapPeers = append(bootstrapPeers, *peerinfo)
	}

	var options []dht.Option = []dht.Option{dht.BootstrapPeers(bootstrapPeers...)}

	// if no bootstrap peers give this peer act as a bootstraping node
	// other peers can use this peers ipfs address for peer discovery via dht
	if mode == "bootstrap" {
		options = append(options, dht.Mode(dht.ModeServer))
	}

	if len(bootstrapPeers) == 0 && mode != "bootstrap" {
		creationLogger.Panic("No bootstrap peers given to the ")
	}

	rettiwtDHT, err := coreDHT.NewKademliaDHT(host, ctx, options...)
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
				creationLogger.Warning("Error connecting to bootstrap node: ", err)
			} else {
				creationLogger.Info("Connection established with bootstrap node: ", bootPeerRef)
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
