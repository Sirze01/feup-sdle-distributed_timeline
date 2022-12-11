package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	contentRouting "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/content-routing"
	kadDHT "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht"
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht/record/userid"
	peer "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/rettiwt-peer"
	postretrieval "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/rettiwt-peer/post-retrieval"
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/timeline"
	"github.com/ipfs/go-log"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	libp2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/procyon-projects/chrono"
)

var dht *kadDHT.KademliaDHT
var host *libp2phost.Host
var timelines []*timeline.UserTimeline
var personalTimeline *timeline.UserTimeline

var identityFilePath *string

var globalContext *context.Context
var globalpubsub *pubsub.PubSub

func setupHandlers() {
	http.HandleFunc("/register", registerHandler)

	http.HandleFunc("/login", func(reply http.ResponseWriter, request *http.Request) {
		loginHandler(reply, request)
		request.URL.Query().Get("username")
		setupUser(request.URL.Query().Get("username"))
	})

	http.HandleFunc("/publish", publishHandler)

	http.HandleFunc("/follow", followHandler)

	http.HandleFunc("/unfollow", unfollowHandler)

	http.HandleFunc("/timeline/all", allTimelineHandler)

	http.HandleFunc("/timeline", timelineHandler)
}

func setupUser(username string) {
	a := *host
	dht.PutValue("/"+userid.UserIDNS+"/"+a.ID().String(), []byte(username))

	_, err := peer.RecordInit(&username, dht, *host) // nodeRecord here
	if err != nil {
		fmt.Println(err)
		return
	}

	pubSub := peer.PubSubInit(*globalContext, *host, username, *identityFilePath)
	globalpubsub = pubSub
	postretrieval.RegisterProtocolHandler(*host, &timelines)
	timelines, personalTimeline = timeline.StartTimelines(username, dht, pubSub, *globalContext, a.ID(), *identityFilePath)

	timeline.CacheCleaner(timelines)

	taskScheduler := chrono.NewDefaultTaskScheduler()

	_, err = taskScheduler.ScheduleWithFixedDelay(func(ctx context.Context) {
		contentRouting.UpdateTimeline(*globalContext, dht, *host, timelines, personalTimeline.Owner, *identityFilePath)
	}, 5*time.Second)

	if err == nil {
		fmt.Println("Task has been scheduled successfully.")
	}
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	globalContext = &ctx

	hosta, dhta := peer.NodeInit(ctx, *identityFilePath, *bootstrapPeersListFilePath, *backendPort)
	host = &hosta
	dht = dhta
	setupHandlers()
	err := http.ListenAndServe(fmt.Sprintf(":%d", *frontendPort), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
}
