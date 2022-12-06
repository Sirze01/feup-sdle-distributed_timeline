package account

var AccountNS = "/account/"

type AccountNSValidator struct{}

func (validator AccountNSValidator) Validate(key string, value []byte) error {
	return nil
}

func (validator AccountNSValidator) Select(key string, values [][]byte) (int, error) {
	return 0, nil
}
