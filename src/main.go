package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	log "github.com/ipfs/go-log/v2"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/bootstrap"
	contentRouting "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/content-routing"
	peer "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/rettiwt-peer"
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/timeline"
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
	case "peer":
		if *username == "" || *password == "" {
			fmt.Println("missing username or password")
			flag.Usage()
			return
		}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		host, dht := peer.NodeInit(ctx, *identityFilePath, *bootstrapPeersListFilePath, *port)

		err := peer.RegisterUser(*register, dht, *username, *password)
		if err != nil {
			fmt.Println(err)
			return
		}

		err = peer.LoginUser(dht, *username, *password)
		if err != nil {
			fmt.Println(err)
			return
		}

		_, err = peer.RecordInit(username, dht, host) // nodeRecord here
		if err != nil {
			fmt.Println(err)
			return
		}

		var timelines []*timeline.UserTimeline
		var personalTimeline *timeline.UserTimeline

		pubSub := peer.PubSubInit(ctx, host, *username, *identityFilePath)

		timelines, personalTimeline = timeline.StartTimelines(*username, dht, pubSub, ctx, host.ID(), *identityFilePath)

		var text string

		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Please enter command (help | publish <string> | followers <string> | follow <string> | unfollow <string> | update) : ")
			text, _ = reader.ReadString('\n')
			text = strings.Replace(text, "\n", "", -1)
			words := strings.Fields(text)
			switch words[0] {
			case "publish":
				cid := contentRouting.NewCID(personalTimeline, host.ID().String())
				personalTimeline.NewPost(cid, words[1])
				contentRouting.ProvideNewPost(cid, dht)
				contentRouting.AnounceNewPost(personalTimeline, *cid)

			// case "followers":
			// 	timeline.GetFollowers(timelines, dht, words[1])
			case "follow":
				// Ask dht for history
				// Ask dht for providers for each post cid -> Get them and annouce ourselves as providers of them
				// Follow the user pubsub topic
				timelines = timeline.FollowUser(timelines, pubSub, ctx, host.ID(), *username, words[1])

			case "unfollow":
				timelines = timeline.UnfollowUser(timelines, words[1])

			case "update":
				// On message from pubsub topic, ask dht for providers of the post cid -> Get it and annouce ourselves as providers of it
				timeline.UpdateTimeline(timelines) // Gets all the pending posts for each subscribed timeline
				for _, timeline := range timelines {
					for _, post := range timeline.PendingPosts {
						addr, _ := dht.FindProviders(*post)
						fmt.Println(addr)
						break
					}
					break
				}

				// Get the posts

				// Provide the posts

			case "help":
				fmt.Println("publish <string> - Publishes a tweet")
				fmt.Println("follow <string> - Follows a user")
				fmt.Println("unfollow <string> - Unfollows a user")
				fmt.Println("update - Updates the timeline")
			default:
				fmt.Println("Invalid command")
				// var connpeers, apppeers string
				// for _, peer := range myTopic.ListPeers() {
				// 	connpeers += peer.String() + " "
				// }
				// fmt.Println(connpeers)
				// for _, peer := range appTopic.ListPeers() {
				// 	apppeers += peer.String() + " "
				// }
				// fmt.Println(apppeers)
			}

			timeline.SaveTimelinesAndPosts(timelines, *username, *identityFilePath)

		}

	default:
		fmt.Println("Error: invalid mode")
		flag.Usage()
		return
	}
}
