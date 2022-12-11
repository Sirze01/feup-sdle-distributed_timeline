package postretrieval

import (
	"encoding/json"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/timeline"
	"github.com/ipfs/go-cid"

	"github.com/libp2p/go-libp2p/core/network"
)

func readRequest(stream network.Stream) (*cid.Cid, error) {
	decoder := json.NewDecoder(stream)
	var request []byte
	err := decoder.Decode(&request)
	if err != nil {
		return nil, err
	}

	var requestedCid cid.Cid
	err = requestedCid.UnmarshalJSON(request)
	return &requestedCid, err
}

func postRetrievalHandler(stream network.Stream) {
	requestedCid, err := readRequest(stream)
	if err != nil {
		logger.Error(err)
		stream.Reset()
	}

	logger.Info("Post request received for CID: ", requestedCid.String())
	post, err := timeline.RetrievePostFromCid(*requestedCid, *timelines)
	if err != nil {
		logger.Error(err)
		stream.Reset()
	}

	encoder := json.NewEncoder(stream)
	err = encoder.Encode(post)
	if err != nil {
		logger.Error(err)
	}
}
