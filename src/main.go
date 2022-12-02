package main

import (
	"flag"
	"fmt"
)

func main() {
	flag.Usage = func() {
		fmt.Println("Usage: rettiwt -m (bootstrap -i idFilePath)|(peer [-r] -u username -w password) [options]")
		fmt.Println("Options:\n\t-p port")
	}

	mode := flag.String("m", "peer", "bootstrap or peer")
	idFilePath := flag.String("i", "", "bootstrap node ID file path")
	register := flag.Bool("r", false, "register a new user")
	username := flag.String("u", "", "username")
	password := flag.String("w", "", "password")

	port := flag.Int("p", 8080, "port")

	flag.Parse()

	if *mode == "bootstrap" {
		if *idFilePath == "" {
			flag.Usage()
			fmt.Println("Error: bootstrap node ID file path is required")
			return
		}

		bootstrapNodeInit(*idFilePath, *port)
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
