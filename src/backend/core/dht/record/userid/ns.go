package userid

var UserIDNS = "userid"

type UserIDNSValidator struct{}

func (validator UserIDNSValidator) Validate(key string, value []byte) error {
	return nil
}

func (validator UserIDNSValidator) Select(key string, values [][]byte) (int, error) {
	return 0, nil
}
