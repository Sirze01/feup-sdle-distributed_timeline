package dht

import (
	"context"
	"fmt"

	recordAccount "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht/record/account"
	recordpeer "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht/record/rettiwt-peer"
	kad "github.com/libp2p/go-libp2p-kad-dht"
	recordlibp2p "github.com/libp2p/go-libp2p-record"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/routing"
)

type KademliaDHT struct {
	IpfsDHT *kad.IpfsDHT
	ctx     context.Context
}

func NewKademliaDHT(host host.Host, ctx context.Context, options ...kad.Option) (*KademliaDHT, error) {
	ipfsDHT, err := kad.New(ctx, host, options...)

	ipfsDHT.Validator.(recordlibp2p.NamespacedValidator)[recordAccount.AccountNS] = recordAccount.AccountNSValidator{}
	ipfsDHT.Validator.(recordlibp2p.NamespacedValidator)[recordpeer.RettiwtPeerNS] = recordpeer.RettiwtPeerNSValidator{}

	if err != nil {
		return nil, err
	}

	return &KademliaDHT{ipfsDHT, ctx}, nil
}

func (kadDHT KademliaDHT) KeyExists(key string) (bool, error) {
	_, err := kadDHT.IpfsDHT.GetValue(kadDHT.ctx, key)

	if err == routing.ErrNotFound {
		return false, nil
	} else if err != nil {
		return false, err
	}

	return true, nil
}

func (kadDHT KademliaDHT) GetValue(key string) ([]byte, error) {
	val, err := kadDHT.IpfsDHT.GetValue(kadDHT.ctx, key)

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

	fmt.Println("Putting value in DHT")

	err = kadDHT.IpfsDHT.PutValue(kadDHT.ctx, key, value)

	return oldVal, err
}
