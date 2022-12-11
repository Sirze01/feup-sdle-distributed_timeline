package contentRouting

import (
	"context"
	"fmt"
	"time"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht"
	peerns "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht/record/rettiwt-peer"
	postretrieval "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/rettiwt-peer/post-retrieval"
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/timeline"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/host"
)

func UpdateTimeline(ctx context.Context, dht dht.ContentProvider, host host.Host, timelines []*timeline.UserTimeline, username, identityFilePath string) {
	allPendingPosts := []*cid.Cid{}

	fmt.Println("Updating timeline...")
	for _, currTimeline := range timelines {
		newPendingPosts := []*cid.Cid{}
		marshaledPeerRecord, _ := dht.GetValue("/" + peerns.RettiwtPeerNS + "/" + currTimeline.Owner) // TODO: Handle error

		peerRecord := PeerRecordUnmarshalJson(marshaledPeerRecord)
		for _, cidRecord := range peerRecord.CidsCache {
			if !cidRecord.ExpireDate.After(time.Now()) {
				continue
			}

			_, ok := currTimeline.Posts[cidRecord.Cid.String()]
			if ok {
				continue
			}

			newPendingPosts = append(newPendingPosts, &cidRecord.Cid)
		}
		// Ask dht for providers for each post cid -> Get them and annouce ourselves as providers of them
		// Follow the user pubsub topic

		for _, cid := range newPendingPosts {
			addr, _ := dht.FindProviders(*cid)
			fmt.Println(addr)

			for _, peer := range addr {
				post, err := postretrieval.RetrievePost(ctx, host, peer, *cid)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println(post)
				currTimeline.Posts[cid.String()] = *post
				ProvideNewPost(cid, dht, currTimeline.Owner)
				break
			}
		}
	}
	timeline.SaveTimelinesAndPosts(timelines, username, identityFilePath)

	for _, timeline := range timelines {
		allPendingPosts = append(allPendingPosts, timeline.PendingPosts...)
	}

	// TODO: Log instead of printing
	fmt.Println("CIDS:")
	for _, cid := range allPendingPosts {
		fmt.Println(cid.String())
	}

}
