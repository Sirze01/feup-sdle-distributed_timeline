package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht"
	peer "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/rettiwt-peer"
	postretrieval "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/rettiwt-peer/post-retrieval"
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/timeline"

	"github.com/ipfs/go-log"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/host"
)

var ctx, cancel = context.WithCancel(context.Background())
var peerHost host.Host = nil
var peerDHT *dht.KademliaDHT = nil
var pubSub *pubsub.PubSub = nil
var loggedIn bool = false

var timelines []*timeline.UserTimeline
var personalTimeline *timeline.UserTimeline
var username string
var identityFilePath *string

func setupRoutes() {
	http.HandleFunc("/helloWorld", func(w http.ResponseWriter, r *http.Request) {
		list := []string{r.RequestURI}
		a, _ := json.Marshal(list)
		w.Write(a)
	})

	// /register?username=USERNAME&password=PASSWORD
	http.HandleFunc("/register", registerHandler)

	// /login?username=USERNAME&password=PASSWORD
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		loginHandler(w, r)
		setupUser()
	})

	// /publish Body: CONTENT
	http.HandleFunc("/publish", publishHandler)

	// /follow?usernameToFollow=USERNAME
	http.HandleFunc("/follow", followHandler)

	// /follow?usernameToUnfollow=USERNAME
	http.HandleFunc("/unfollow", unfollowHandler)

	// /update
	http.HandleFunc("/update", updateHandler)

	// /timeline/all
	http.HandleFunc("/timeline/all", allTimelineHandler)

	// /timeline?username=USERNAME
	http.HandleFunc("/timeline", timelineHandler)

}

func setupNode(identityFilePath, bootstrapPeersListFilePath *string, port *int) {
	fmt.Println("Setting up node...")

	peerHost, peerDHT = peer.NodeInit(ctx, *identityFilePath, *bootstrapPeersListFilePath, *port)
}

func setupUser() {
	_, err := peer.RecordInit(&username, peerDHT, peerHost) // nodeRecord here
	if err != nil {
		fmt.Println(err)
		return
	}

	pubSub = peer.PubSubInit(ctx, peerHost, username, *identityFilePath)
	postretrieval.RegisterProtocolHandler(peerHost, &timelines)
	timelines, personalTimeline = timeline.StartTimelines(username, peerDHT, pubSub, ctx, peerHost.ID(), *identityFilePath)
}

func main() {
	fmt.Println("Distributed Chat App v0.01")

	flag.Usage = func() {
		fmt.Printf("Usage: %s -i identityFilePath -l bootstrapPeersListFilePath [options]", os.Args[0])
		fmt.Println("Optional:\n\t-p port\n\t--log logLevel\n\t--usage")

		flag.PrintDefaults()
	}

	// Common arguments
	identityFilePath = flag.String("i", "", "bootstrap node ID file path")
	bootstrapPeersListFilePath := flag.String("l", "", "bootstrap nodes IDs list file path")

	// Optional arguments
	backendPort := flag.Int("b", 7001, "backend port")
	frontendPort := flag.Int("f", 8001, "frontend port")

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
	defer cancel()
	setupNode(identityFilePath, bootstrapPeersListFilePath, backendPort)
	setupRoutes()
	http.ListenAndServe(fmt.Sprintf(":%d", *frontendPort), nil)
	select {}
}
