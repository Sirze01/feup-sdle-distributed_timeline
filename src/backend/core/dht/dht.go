package dht

import "github.com/libp2p/go-libp2p-core/peer"

type DHT interface {
	KeyExists(key string) (bool, error)
	GetValue(key string) ([]byte, error)
	PutValue(key string, value []byte) ([]byte, error)
	GetPeerID() peer.ID
}
