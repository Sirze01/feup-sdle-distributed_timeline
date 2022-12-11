package peer

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	log "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
)

var idLogger = log.Logger("identity")

type KeyPair struct {
	PrvKey []byte
	PubKey []byte
}

func CreateIdentity(idFilePath string) (crypto.PrivKey, error) {
	var prvKey crypto.PrivKey
	var pubKey crypto.PubKey

	idLogger.Info("No ID file found. Generating new key pair")
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	prvKey, pubKey, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)

	if err != nil {
		idLogger.Panic(err)
	}

	idLogger.Info("Saving key pair to file")
	if saveIdentity(idFilePath, prvKey, pubKey) != nil {
		idLogger.Panic(err)
	}

	return prvKey, nil
}

func LoadIdentity(idFilePath string) (crypto.PrivKey, error) {
	idLogger.Info("ID file found. Loading key pair")
	prvKey, _, err := loadIdentity(idFilePath)

	if err != nil {
		idLogger.Panic(err)
	}
	return prvKey, nil
}

func loadIdentity(idFilePath string) (crypto.PrivKey, crypto.PubKey, error) {
	var pair KeyPair
	jsonPair, err := os.ReadFile(idFilePath)
	if err != nil {
		return nil, nil, err
	}

	err = json.Unmarshal(jsonPair, &pair)
	if err != nil {
		return nil, nil, err
	}

	prvKey, err := crypto.UnmarshalPrivateKey(pair.PrvKey)
	if err != nil {
		return nil, nil, err
	}

	pubKey, err := crypto.UnmarshalPublicKey(pair.PubKey)
	if err != nil {
		return nil, nil, err
	}
	return prvKey, pubKey, nil
}

func saveIdentity(idFilePath string, prvKey crypto.PrivKey, pubKey crypto.PubKey) error {
	prvKeyBytes, err := crypto.MarshalPrivateKey(prvKey)
	if err != nil {
		return err
	}
	pubKeyBytes, err := crypto.MarshalPublicKey(pubKey)
	if err != nil {
		return err
	}

	pair := KeyPair{
		PrvKey: prvKeyBytes,
		PubKey: pubKeyBytes,
	}

	jsonPair, err := json.MarshalIndent(pair, "", "    ")
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(idFilePath), 0777)
	if err != nil {
		return err
	}

	err = os.WriteFile(idFilePath, jsonPair, 0666)
	if err != nil {
		return err
	}
	return nil
}
