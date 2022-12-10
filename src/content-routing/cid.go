package contentRouting

//package contentRouting

import (
	"time"

	"github.com/ipfs/go-cid"
	u "github.com/ipfs/go-ipfs-util"
)

const DEFAULT_RECORD_DURATION = 24

type PostCIDRecord struct {
	Cid        cid.Cid
	ExpireDate time.Time
}

func NewPostCID(postID string) cid.Cid {
	return cid.NewCidV1(cid.DagCBOR, u.Hash([]byte(postID)))
}

func NewPostCIDRecord(cid cid.Cid, durationHours int) PostCIDRecord {
	var duration time.Duration
	if durationHours == 0 {
		duration = time.Duration(DEFAULT_RECORD_DURATION) * time.Hour
	} else {
		duration = time.Duration(durationHours) * time.Hour
	}

	return PostCIDRecord{
		Cid:        cid,
		ExpireDate: time.Now().Add(duration),
	}
}
