package contentRouting

import (
	"encoding/json"
)

type RettiwtPeerRecord struct {
	Username  string
	CidsCache []MessageCIDRecord
}

func PeerRecordMarshalJson(record *RettiwtPeerRecord) []byte {
	marshaledRecord, _ := json.Marshal(*record)
	return marshaledRecord
}

func PeerRecordUnmarshalJson(marshaledRecord []byte) *RettiwtPeerRecord {
	record := new(RettiwtPeerRecord)
	json.Unmarshal(marshaledRecord, record)
	return record
}

func (record RettiwtPeerRecord) addCID(cidRecord MessageCIDRecord) {
	record.CidsCache = append(record.CidsCache, cidRecord)
}

func (record RettiwtPeerRecord) deleteCID(i int) {
	record.CidsCache[i] = record.CidsCache[len(record.CidsCache)-1]

	record.CidsCache = record.CidsCache[:len(record.CidsCache)-1]
}
