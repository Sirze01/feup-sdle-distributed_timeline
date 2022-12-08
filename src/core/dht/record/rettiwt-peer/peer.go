package peer

import "encoding/json"

var RettiwtPeerNS = "rettiwt-peer"

type RettiwtPeerRecord struct {
	Username string
}

func MarshalJson(record *RettiwtPeerRecord) []byte {
	marshaledRecord, _ := json.Marshal(*record)
	return marshaledRecord
}

func UnmarshalJson(marshaledRecord []byte) *RettiwtPeerRecord {
	record := new(RettiwtPeerRecord)
	json.Unmarshal(marshaledRecord, record)
	return record
}

type RettiwtPeerNSValidator struct{}

func (validator RettiwtPeerNSValidator) Validate(key string, value []byte) error {
	return nil
}

func (validator RettiwtPeerNSValidator) Select(key string, values [][]byte) (int, error) {
	return 0, nil
}
