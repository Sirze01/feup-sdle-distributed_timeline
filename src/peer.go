package main

import (
	"context"
	"encoding/json"
	"math/rand"
	"os"
	filepath "path/filepath"
	"time"

	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
)

func peerNodeInit(register, username, password string, port int) {
	ctx := context.Background()

	var prvKey crypto.PrivKey
	var pubKey crypto.PubKey
	var idFilePath = "./nodes/" + username + ".json"
	if (register){
		if _, err := os.Stat(idFilePath); err == nil {
			// Generate a new key pair for this host. We will use it to obtain a valid host ID.
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			prvKey, pubKey, err = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)

			if err != nil {
				panic(err)
			}

			//check if username already in use
			host, err := libp2p.New(libp2p.Identity(prvKey))
			if err != nil {
				panic(err)
			}

			_, err = dht.New(ctx, host)
			if err != nil {
				panic(err)
			}

			//send message to dht with the username and check if it is already in use


			//if username is not in use, save the keypair and create a new account

			if saveNodeState(idFilePath, prvKey, pubKey) != nil {
				panic(err)
			}

			//else, return error

			panic(err)
		} else {
			panic(err)
			
		}
	} else {
		prvKey, pubKey, err := loadNodeState(idFilePath)
		if err != nil {
			panic(err)
		}
	}

	
	// if _, err := os.Stat(idFilePath); err != nil {
	// 	// Generate a new key pair for this host. We will use it to obtain a valid host ID.
	// 	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 	prvKey, pubKey, err = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)

	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	if saveKeyPair(idFilePath, prvKey, pubKey) != nil {
	// 		panic(err)
	// 	}
	// } else {
	// 	prvKey, _, err = loadKeyPair(idFilePath)

	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

	// host, err := libp2p.New(libp2p.Identity(prvKey))
	// if err != nil {
	// 	panic(err)
	// }

	// _, err = dht.New(ctx, host)
	// if err != nil {
	// 	panic(err)
	// }

	// select {}

}

func loadNodeState(idFilePath string) (crypto.PrivKey, crypto.PubKey, error) {
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

func saveNodeState(idFilePath string, prvKey crypto.PrivKey, pubKey crypto.PubKey) error {
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

	json_pair, err := json.MarshalIndent(pair, "", "    ")
	if err != nil {
		return err
	}

	err = os.MkdirAll(filepath.Dir(idFilePath), 0777)
	if err != nil {
		return err
	}

	err = os.WriteFile(idFilePath, json_pair, 0666)
	if err != nil {
		return err
	}
	return nil
}

