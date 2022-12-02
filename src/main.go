package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: rettiwt -m (bootstrap -i idFilePath)|(peer -u username -w password) -list idsList [options]")
		fmt.Println("Options:\n\t-p port")
	}

	mode := flag.String("m", "peer", "bootstrap or peer")
	idFilePath := flag.String("i", "", "bootstrap node ID file path")
	idsListFilePath := flag.String("list", "", "bootstrap nodes IDs list file path")
	username := flag.String("u", "", "username")
	password := flag.String("w", "", "password")
	port := flag.Int("p", 8080, "port")

	flag.Parse()

	if *idsListFilePath == "" {
		flag.Usage()
		fmt.Println("Error: missing bootstrap nodes IDs list file path")
		return
	}

	if *mode == "bootstrap" {
		if *idsListFilePath == "" {
			flag.Usage()
			fmt.Println("Error: bootstrap node ID file path is required")
			return
		}

		bootstrapNodeInit(*idFilePath, *idsListFilePath, *port)
		fmt.Printf("bootstrap mode on port %d\n", *port)
	} else if *mode == "peer" {
		if *username == "" || *password == "" {
			flag.Usage()
			return
		}
		peerNodeInit(*username, *password, *port)
		fmt.Printf("peer mode on port %d, username %s, password %s\n", *port, *username, *password)
	} else {
		flag.Usage()
		return
	}

	// start libp2p node peer or bootstrap node
	// ( Save the node to a file )
	// bootstrap to the network
	// protocol handlers -> co routine?
	// protocol function sending

	// Posts Get/Send
	// Follow/Unfollow system
}
