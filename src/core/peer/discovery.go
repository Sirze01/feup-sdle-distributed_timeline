package peer

import (
	"context"
	"fmt"
	"time"

	coreDHT "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht"
	"github.com/libp2p/go-libp2p-core/network"
	discovery "github.com/libp2p/go-libp2p-discovery"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/p2p/discovery/routing"
	"github.com/libp2p/go-libp2p/p2p/discovery/util"
)

func DiscoverPeers(ctx context.Context, h host.Host, dht *coreDHT.KademliaDHT, announceString string) {
	routingDiscovery := routing.NewRoutingDiscovery(dht.IpfsDHT)
	util.Advertise(ctx, routingDiscovery, announceString)

	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:

			peers, err := discovery.FindPeers(ctx, routingDiscovery, announceString)
			if err != nil {
				creationLogger.Panic(err)
			}

			for _, p := range peers {
				if p.ID == h.ID() {
					continue
				}
				if h.Network().Connectedness(p.ID) != network.Connected {
					_, err = h.Network().DialPeer(ctx, p.ID)
					fmt.Printf("Connected to peer %s\n", p.ID.String())
					if err != nil {
						continue
					}
				}
			}
		}
	}
}
