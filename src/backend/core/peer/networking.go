package peer

import (
	log "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"
)

var networkLogger = log.Logger("network")

func GetNotifiee() network.NotifyBundle {
	var networkNotifiee network.NotifyBundle
	networkNotifiee.ListenF = func(net network.Network, ma multiaddr.Multiaddr) {
		networkLogger.Infof("Listening on %s, on interface %s", ma, net)
	}

	networkNotifiee.ConnectedF = func(net network.Network, con network.Conn) {
		networkLogger.Infof("Connected to %s on interface %s", con, net)
	}
	return networkNotifiee
}
