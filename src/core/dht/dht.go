package dht

type DHT interface {
	KeyExists(key string) (bool, error)
	GetValue(key string) ([]byte, error)
	PutValue(key string, value []byte) ([]byte, error)
}
