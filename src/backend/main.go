package main

import (
	"flag"
	"fmt"
	"os"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/bootstrap"
	peer "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/rettiwt-peer"
	log "github.com/ipfs/go-log/v2"
)

func main() {
	// Define usage function
	flag.Usage = func() {
		fmt.Printf("Usage: %s -m bootstrap|(peer [-r] -u username -w password) -i identityFilePath -l bootstrapPeersListFilePath [options]", os.Args[0])
		fmt.Println("Optional:\n\t-p port\n\t--log logLevel\n\t--usage")

		flag.PrintDefaults()
	}

	// Common arguments
	mode := flag.String("m", "peer", "bootstrap or peer")
	identityFilePath := flag.String("i", "", "bootstrap node ID file path")
	bootstrapPeersListFilePath := flag.String("l", "", "bootstrap nodes IDs list file path")

	// Arguments for peer mode only
	register := flag.Bool("r", false, "register a new user")
	username := flag.String("u", "", "username")
	password := flag.String("w", "", "password")

	// Optional arguments
	port := flag.Int("p", 7001, "port")
	logLevel := flag.String("log", "", "log level")
	usage := flag.Bool("usage", false, "show usage")

	flag.Parse()

	// Show usage if requested
	if *usage {
		flag.Usage()
		return
	}

	// Set log level
	switch *logLevel {
	case "debug":
		log.SetAllLoggers(log.LevelDebug)
	case "error":
		log.SetAllLoggers(log.LevelError)
	case "fatal":
		log.SetAllLoggers(log.LevelFatal)
	case "panic":
		log.SetAllLoggers(log.LevelPanic)
	case "warn":
		log.SetAllLoggers(log.LevelWarn)
	case "info":
		fallthrough
	case "":
		fallthrough
	default:
		log.SetAllLoggers(log.LevelInfo)
	}

	// Check if identity file and bootstrap peers list file path  is provided
	if *identityFilePath == "" {
		fmt.Println("Error: missing node identity file path")
		flag.Usage()
		return
	}

	if *bootstrapPeersListFilePath == "" {
		fmt.Println("Error: missing bootstrap nodes IDs list file path")
		flag.Usage()
		return
	}

	// Delegate peer initialization

	switch *mode {
	case "bootstrap":
		bootstrap.BootstrapNodeInit(*identityFilePath, *bootstrapPeersListFilePath, *port)
		fmt.Printf("bootstrap mode on port %d\n", *port)
	case "peer":
		if *username == "" || *password == "" {
			fmt.Println("missing username or password")
			flag.Usage()
			return
		}
		peer.NodeInit(*identityFilePath, *bootstrapPeersListFilePath, *register, *username, *password, *port)
		fmt.Printf("peer mode on port %d, username %s, password %s\n", *port, *username, *password)
	default:
		fmt.Println("Error: invalid mode")
		flag.Usage()
		return
	}
}
