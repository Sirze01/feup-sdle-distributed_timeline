package contentRouting

//package contentRouting

import (
	"strconv"
	"time"

	"github.com/ipfs/go-cid"
	u "github.com/ipfs/go-ipfs-util"
)

const DEFAULT_RECORD_DURATION = 24

type MessageCIDRecord struct {
	Cid        cid.Cid
	ExpireDate time.Time
}

func NewMessageCID(messageID int) cid.Cid {
	return cid.NewCidV1(cid.DagCBOR, u.Hash([]byte(strconv.Itoa(messageID))))
}

func NewMessageCIDRecord(cid cid.Cid, durationHours int) MessageCIDRecord {
	var duration time.Duration
	if durationHours == 0 {
		duration = time.Duration(DEFAULT_RECORD_DURATION) * time.Hour
	} else {
		duration = time.Duration(durationHours) * time.Hour
	}

	return MessageCIDRecord{
		Cid:        cid,
		ExpireDate: time.Now().Add(duration),
	}
}
