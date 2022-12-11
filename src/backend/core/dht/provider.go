package dht

import (
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
)

type ContentProvider interface {
	DHT
	Provide(cid.Cid) error
	FindProviders(cid.Cid) ([]peer.AddrInfo, error)
}
