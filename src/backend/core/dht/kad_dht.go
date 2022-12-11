package dht

import (
	"context"
	"fmt"

	recordAccount "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht/record/account"
	peerns "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht/record/rettiwt-peer"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
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
	ipfsDHT.Validator.(recordlibp2p.NamespacedValidator)[peerns.RettiwtPeerNS] = peerns.RettiwtPeerNSValidator{}

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

func (kadDHT KademliaDHT) Provide(cid cid.Cid) error {
	return kadDHT.IpfsDHT.Provide(kadDHT.ctx, cid, true)
}

func (kadDHT KademliaDHT) FindProviders(cid cid.Cid) ([]peer.AddrInfo, error) {
	return kadDHT.IpfsDHT.FindProviders(kadDHT.ctx, cid)
}

func (kadDHT KademliaDHT) GetPeerID() peer.ID {
	return kadDHT.IpfsDHT.Host().ID()
}
