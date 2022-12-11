package postretrieval

import (
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/timeline"
	log "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/protocol"
)

var logger = log.Logger("post-retrieval")

var protocolName = "post-retrieval"
var protocolVersion = "0.1"

var timelines *[]*timeline.UserTimeline

func getProtoId() protocol.ID {
	return protocol.ID(protocolName + "/" + protocolVersion)
}

func RegisterProtocolHandler(host host.Host, nodeTimelines *[]*timeline.UserTimeline) {
	timelines = nodeTimelines
	host.SetStreamHandler(getProtoId(), postRetrievalHandler)
}
