package contentRouting

import (
	"context"
	"log"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht"
	peerns "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht/record/rettiwt-peer"
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/timeline"
	"github.com/ipfs/go-cid"
	"github.com/procyon-projects/chrono"
)

func AnounceNewMessage(timeline *timeline.ChatRoom) (*cid.Cid, error) {
	timeline.CurrMessageID += 1

	cid := NewMessageCID(timeline.CurrMessageID)

	cidMarshaled, _ := cid.MarshalJSON()

	timeline.Publish(string(cidMarshaled))

	return &cid, nil
}

func ProvideNewMessage(cid *cid.Cid, dht dht.ContentProvider, timeline *timeline.ChatMessage) error {
	err := dht.Provide(*cid)

	if err != nil {
		return err
	}

	marshaledPeerRecord, err := dht.GetValue("/" + peerns.RettiwtPeerNS + "/" + dht.GetPeerID().String())

	peerRecord := PeerRecordUnmarshalJson(marshaledPeerRecord)

	cidRecord := NewMessageCIDRecord(*cid, 0)

	peerRecord.addCID(cidRecord)

	SetCIDDeleteHandler(&cidRecord, peerRecord, dht)

	return nil
}

func SetCIDDeleteHandler(cidRecord *MessageCIDRecord, peerRecord *RettiwtPeerRecord, dht dht.ContentProvider) error {
	taskScheduler := chrono.NewDefaultTaskScheduler()

	_, err := taskScheduler.Schedule(func(ctx context.Context) {
		log.Print("One-Shot Task")

		for i, cidRecord2 := range peerRecord.CidsCache {
			if cidRecord.Cid.Equals(cidRecord2.Cid) {

				peerRecord.deleteCID(i)

				marshaledPeerRecord := PeerRecordMarshalJson(peerRecord)

				dht.PutValue("/"+peerns.RettiwtPeerNS+"/"+dht.GetPeerID().String(), marshaledPeerRecord)
			}
		}
	}, chrono.WithTime(cidRecord.ExpireDate))

	return err
}
