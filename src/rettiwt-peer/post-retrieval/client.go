package postretrieval

import (
	"context"
	"encoding/json"
	"errors"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/timeline"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

func readReply(stream network.Stream) (*timeline.TimelinePost, error) {
	decoder := json.NewDecoder(stream)
	var reply timeline.TimelinePost
	err := decoder.Decode(&reply)
	if err != nil {
		return nil, err
	}
	return &reply, nil
}

func RetrievePost(ctx context.Context, host host.Host, peer peer.AddrInfo, cid cid.Cid) (*timeline.TimelinePost, error) {
	if peer.ID == host.ID() {
		return nil, errors.New("can't retrieve from self")
	}

	host.Connect(ctx, peer)

	stream, err := host.NewStream(ctx, peer.ID, getProtoId())
	if err != nil {
		logger.Error("can't open stream: ", err)
		stream.Reset()
		return nil, err
	}

	request, err := cid.MarshalJSON()
	if err != nil {
		logger.Error(err)
		stream.Reset()
		return nil, err
	}

	encoder := json.NewEncoder(stream)
	err = encoder.Encode(request)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	replyPost, err := readReply(stream)
	if err != nil {
		logger.Error(err)
		stream.Reset()
		return nil, err
	}
	return replyPost, nil
}
