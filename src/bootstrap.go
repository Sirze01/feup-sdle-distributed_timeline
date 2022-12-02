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
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/multiformats/go-multiaddr"

	log "github.com/ipfs/go-log/v2"
)

var logger = log.Logger("bootstrap")

func bootstrapNodeInit(idFilePath, idsListFilePath string, port int) {
	log.SetLogLevel("bootstrap", "info")

	ctx := context.Background()
	var prvKey crypto.PrivKey
	var pubKey crypto.PubKey

	if _, err := os.Stat(idFilePath); err != nil {
		// Generate a new key pair for this host. We will use it to obtain a valid host ID.
		logger.Info("No ID file found. Generating new key pair")
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		prvKey, pubKey, err = crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)

		if err != nil {
			panic(err)
		}

		// Save it for later use on creating the bootstrap nodes list
		logger.Info("Saving key pair to file")
		if saveKeyPair(idFilePath, prvKey, pubKey) != nil {
			panic(err)
		}
	} else {
		// Load the key pair from the file to boot as a bootstrap node
		logger.Info("ID file found. Loading key pair")
		prvKey, _, err = loadKeyPair(idFilePath)

		if err != nil {
			panic(err)
		}
	}

	var networkNotifiee network.NotifyBundle
	networkNotifiee.ListenF = func(net network.Network, ma multiaddr.Multiaddr) {
		logger.Info("Listening on %s, on interface %s", ma, net)
	}

	networkNotifiee.ConnectedF = func(net network.Network, con network.Conn) {
		logger.Info("Connected to %s on interface %s", con, net)
	}

	// Create a new libp2p Host that uses the provided identity
	host, err := libp2p.New(libp2p.Identity(prvKey))
	if err != nil {
		panic(err)
	}
	host.Network().Notify(&networkNotifiee)
	logger.Infof("Host created. We are: %s", host.ID())
	err = saveNodeId(idsListFilePath, host.ID().String())
	if err != nil {
		logger.Warn("Error saving node ID to list: ", err)
	}

	_, err = dht.New(ctx, host)
	if err != nil {
		panic(err)
	}

	select {}
}

type KeyPair struct {
	PrvKey []byte
	PubKey []byte
}

func loadKeyPair(idFilePath string) (crypto.PrivKey, crypto.PubKey, error) {
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

func saveKeyPair(idFilePath string, prvKey crypto.PrivKey, pubKey crypto.PubKey) error {
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

func saveNodeId(idsListFilePath string, id string) error {
	var ids []string
	var jsonIds []byte
	if _, err := os.Stat(idsListFilePath); err != nil {
		ids = append(ids, id)

		jsonIds, err = json.MarshalIndent(ids, "", "    ")
		if err != nil {
			return err
		}
	} else {
		jsonIds, err = os.ReadFile(idsListFilePath)
		if err != nil {
			return err
		}

		err = json.Unmarshal(jsonIds, &ids)
		if err != nil {
			return err
		}

		for _, listedID := range ids {
			if id == listedID {
				return nil
			}
		}

		ids = append(ids, id)
		jsonIds, err = json.MarshalIndent(ids, "", "    ")
		if err != nil {
			return err
		}
	}

	err := os.MkdirAll(filepath.Dir(idsListFilePath), 0777)
	if err != nil {
		return err
	}

	err = os.WriteFile(idsListFilePath, jsonIds, 0666)

	if err != nil {
		return err
	}

	return nil
}
