package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: rettiwt -m (bootstrap -i idFilePath)|(peer [-r] -u username -w password) -list idsList [options]")
		fmt.Println("Options:\n\t-p port")
	}

	mode := flag.String("m", "peer", "bootstrap or peer")
	idsListFilePath := flag.String("list", "", "bootstrap nodes IDs list file path")

	// For bootstrap mode only
	idFilePath := flag.String("i", "", "bootstrap node ID file path")

	// For peer mode only
	register := flag.Bool("r", false, "register a new user")
	username := flag.String("u", "", "username")
	password := flag.String("w", "", "password")

	// Common options
	port := flag.Int("p", 7001, "port")

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
		peerNodeInit(*register, *username, *password, *port)
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
