package dht

import (
	"context"

	"github.com/libp2p/go-libp2p-core/routing"
	kad "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/host"
)

type KademliaDHT struct {
	ipfsDHT *kad.IpfsDHT
	ctx     context.Context
}

func NewKademliaDHT(host host.Host, ctx context.Context) (*KademliaDHT, error) {
	ipfsDHT, err := kad.New(ctx, host)

	if err != nil {
		return nil, err
	}

	return &KademliaDHT{ipfsDHT, ctx}, nil
}

func (kadDHT KademliaDHT) KeyExists(key string) (bool, error) {
	_, err := kadDHT.ipfsDHT.GetValue(kadDHT.ctx, key)

	if err == routing.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func (kadDHT KademliaDHT) GetValue(key string) ([]byte, error) {
	val, err := kadDHT.ipfsDHT.GetValue(kadDHT.ctx, key)

	if err == routing.ErrNotFound {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	return val, nil
}

func (kadDHT KademliaDHT) PutValue(key string, value []byte) ([]byte, error) {
	oldVal, err := kadDHT.GetValue(key)

	if err != nil {
		return oldVal, err
	}

	err = kadDHT.ipfsDHT.PutValue(kadDHT.ctx, key, value)

	return oldVal, err
}
