package contentRouting

import (
	"context"
	"fmt"
	"log"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht"
	peerns "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht/record/rettiwt-peer"
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/timeline"
	"github.com/ipfs/go-cid"
	"github.com/procyon-projects/chrono"
)

func NewCID(timeline *timeline.UserTimeline, peerId string) *cid.Cid {
	timeline.CurrPostID += 1
	cid := NewPostCID(peerId + fmt.Sprint(timeline.CurrPostID))
	return &cid
}

func AnounceNewPost(timeline *timeline.UserTimeline, cid cid.Cid) error {
	cidMarshaled, _ := cid.MarshalJSON()

	fmt.Println("Anouncing new post: " + string(cid.String()))

	err := timeline.Publish(string(cidMarshaled))
	return err
}

func ProvideNewPost(cid *cid.Cid, dht dht.ContentProvider, username string) error {
	err := dht.Provide(*cid)

	if err != nil {
		return err
	}

	marshaledPeerRecord, err := dht.GetValue("/" + peerns.RettiwtPeerNS + "/" + username) // TODO: Handle error
	if err != nil {
		fmt.Println(err)
	}

	peerRecord := PeerRecordUnmarshalJson(marshaledPeerRecord)

	cidRecord := NewPostCIDRecord(*cid, 0)

	peerRecord.addCID(cidRecord)

	marshaledPeerRecord = PeerRecordMarshalJson(peerRecord)
	dht.PutValue("/"+peerns.RettiwtPeerNS+"/"+username, marshaledPeerRecord)

	SetCIDDeleteHandler(&cidRecord, peerRecord, dht, username)

	return nil
}

func SetCIDDeleteHandler(cidRecord *PostCIDRecord, peerRecord *RettiwtPeerRecord, dht dht.ContentProvider, username string) error {
	taskScheduler := chrono.NewDefaultTaskScheduler()

	_, err := taskScheduler.Schedule(func(ctx context.Context) {
		log.Print("One-Shot Task")

		for i, cidRecord2 := range peerRecord.CidsCache {
			if cidRecord.Cid.Equals(cidRecord2.Cid) {

				peerRecord.deleteCID(i)

				marshaledPeerRecord := PeerRecordMarshalJson(peerRecord)

				dht.PutValue("/"+peerns.RettiwtPeerNS+"/"+username, marshaledPeerRecord)
			}
		}
	}, chrono.WithTime(cidRecord.ExpireDate))

	return err
}
