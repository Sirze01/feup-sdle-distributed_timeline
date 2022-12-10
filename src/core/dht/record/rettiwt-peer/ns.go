package ns

var RettiwtPeerNS = "rettiwt-peer"

type RettiwtPeerNSValidator struct{}

func (validator RettiwtPeerNSValidator) Validate(key string, value []byte) error {
	return nil
}

func (validator RettiwtPeerNSValidator) Select(key string, values [][]byte) (int, error) {
	return 0, nil
}
