package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	filepath "path/filepath"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/peer"
	log "github.com/ipfs/go-log/v2"
	"github.com/multiformats/go-multiaddr"
)

var logger = log.Logger("bootstrap")

func BootstrapNodeInit(idFilePath, bootstrapPeerIdsFilePath string, port int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create the host
	host := peer.CreateHost(ctx, idFilePath, port)

	var _ = peer.InitDHT(ctx, host, bootstrapPeerIdsFilePath)

	saveNodeId(bootstrapPeerIdsFilePath, host.Addrs(), host.ID().String())

	//go peer.DiscoverPeers(ctx, host, dht)

	select {}
}

func saveNodeId(idsListFilePath string, multiaddr []multiaddr.Multiaddr, id string) {
	var bootstrapNodes []string
	var jsonBoostrapNodes []byte

	if _, err := os.Stat(idsListFilePath); err != nil {
		for _, addr := range multiaddr {
			bootstrapNodes = append(bootstrapNodes, fmt.Sprintf("%s/p2p/%s", addr.String(), id))
		}
	} else {
		jsonBoostrapNodes, err = os.ReadFile(idsListFilePath)
		if err != nil {
			logger.Panic(err)
		}

		err = json.Unmarshal(jsonBoostrapNodes, &bootstrapNodes)
		if err != nil {
			logger.Panic(err)
		}

		found := false
		for _, addr := range multiaddr {
			for _, listedAddr := range bootstrapNodes {
				if listedAddr == addr.String() {
					found = true
					logger.Info("Node %s already listed", addr.String())
					break
				}
			}
			if !found {
				bootstrapNodes = append(bootstrapNodes, fmt.Sprintf("%s/p2p/%s", addr.String(), id))
			}
		}
	}

	jsonBoostrapNodes, err := json.MarshalIndent(bootstrapNodes, "", "    ")
	if err != nil {
		logger.Panic(err)
	}

	err = os.MkdirAll(filepath.Dir(idsListFilePath), 0777)
	if err != nil {
		logger.Panic(err)
	}

	err = os.WriteFile(idsListFilePath, jsonBoostrapNodes, 0666)

	if err != nil {
		logger.Panic(err)
	}
}
