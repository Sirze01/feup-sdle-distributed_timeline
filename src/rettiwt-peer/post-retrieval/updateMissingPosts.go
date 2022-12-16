package postretrieval

import (
	"context"
	"fmt"
	"time"

	contentRouting "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/content-routing"
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht"
	peerns "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht/record/rettiwt-peer"
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/timeline"
	"github.com/libp2p/go-libp2p/core/host"
)

func UpdateTimelineMissingPosts(currTimeline *timeline.UserTimeline, dht dht.ContentProvider, host host.Host, ctx context.Context) {
	marshaledPeerRecord, _ := dht.GetValue("/" + peerns.RettiwtPeerNS + "/" + currTimeline.Owner) // TODO: Handle error

	peerRecord := contentRouting.PeerRecordUnmarshalJson(marshaledPeerRecord)
	for _, cidRecord := range peerRecord.CidsCache {
		if !cidRecord.ExpireDate.After(time.Now()) {
			continue
		}

		addr, _ := dht.FindProviders(cidRecord.Cid)
		fmt.Println(addr)

		for _, peer := range addr {
			post, err := RetrievePost(ctx, host, peer, cidRecord.Cid)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(post)
			currTimeline.Posts[cidRecord.Cid.String()] = *post
			contentRouting.ProvideNewPost(&cidRecord.Cid, dht, currTimeline.Owner)
			break
		}
	}
}
