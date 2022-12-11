package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	contentRouting "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/content-routing"
	peerns "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht/record/rettiwt-peer"
	peer "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/rettiwt-peer"
	postretrieval "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/rettiwt-peer/post-retrieval"
	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/timeline"

	"github.com/ipfs/go-cid"
	"golang.org/x/exp/slices"
)

func registerHandler(reply http.ResponseWriter, request *http.Request) {
	fmt.Println("Registering user")
	if loggedIn {
		bytes, _ := json.Marshal("Node already used")
		reply.Write(bytes)
		return
	}

	query := request.URL.Query()

	err := peer.RegisterUser(true, peerDHT, query.Get("username"), query.Get("password"))
	if err != nil {
		fmt.Println(err)
		return
	}

	reply.WriteHeader(http.StatusOK)
}

func loginHandler(reply http.ResponseWriter, request *http.Request) {
	if loggedIn {
		bytes, _ := json.Marshal("Node already used")
		reply.Write(bytes)
		return
	}

	query := request.URL.Query()

	err := peer.LoginUser(peerDHT, query.Get("username"), query.Get("password"))
	if err != nil {
		bytes, _ := json.Marshal(err)
		reply.WriteHeader(http.StatusBadRequest)
		reply.Write(bytes)
		return
	}

	loggedIn = true
	username = query.Get("username")
	reply.WriteHeader(http.StatusOK)
}

func publishHandler(reply http.ResponseWriter, request *http.Request) {
	if !loggedIn {
		bytes, _ := json.Marshal("Not logged in")
		reply.WriteHeader(http.StatusUnauthorized)
		reply.Write(bytes)
		return
	}

	if request.Method != "POST" {
		bytes, _ := json.Marshal("Not a POST request")
		reply.WriteHeader(http.StatusBadRequest)
		reply.Write(bytes)
		return
	}

	var bodyBytes []byte
	if request.Body != nil {
		bodyBytes, _ = io.ReadAll(request.Body)
	}
	// Restore the io.ReadCloser to its original state
	request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var post string
	json.Unmarshal(bodyBytes, &post)

	cid := contentRouting.NewCID(personalTimeline, peerHost.ID().String())
	personalTimeline.NewPost(cid, post)
	contentRouting.ProvideNewPost(cid, peerDHT, username)
	contentRouting.AnounceNewPost(personalTimeline, *cid)

	reply.WriteHeader(http.StatusOK)
	timeline.SaveTimelinesAndPosts(timelines, username, *identityFilePath)

}

func followHandler(reply http.ResponseWriter, request *http.Request) {
	if !loggedIn {
		bytes, _ := json.Marshal("Not logged in")
		reply.WriteHeader(http.StatusUnauthorized)
		reply.Write(bytes)
		return
	}

	query := request.URL.Query()

	if query.Get("usernameToFollow") == username {
		bytes, _ := json.Marshal("Can't follow yourself")
		reply.WriteHeader(http.StatusBadRequest)
		reply.Write(bytes)
		return
	}
	// Ask dht for history
	_, currTimeline := timeline.FollowUser(&timelines, pubSub, ctx, peerHost.ID(), query.Get("usernameToFollow"))

	marshaledPeerRecord, _ := peerDHT.GetValue("/" + peerns.RettiwtPeerNS + "/" + currTimeline.Owner) // TODO: Handle error

	peerRecord := contentRouting.PeerRecordUnmarshalJson(marshaledPeerRecord)
	for _, cidRecord := range peerRecord.CidsCache {
		if !cidRecord.ExpireDate.After(time.Now()) {
			continue
		}

		addr, _ := peerDHT.FindProviders(cidRecord.Cid)
		fmt.Println(addr)

		for _, peer := range addr {
			post, err := postretrieval.RetrievePost(ctx, peerHost, peer, cidRecord.Cid)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(post)
			currTimeline.Posts[cidRecord.Cid.String()] = *post
			contentRouting.ProvideNewPost(&cidRecord.Cid, peerDHT, currTimeline.Owner)
			break
		}
	}

	reply.WriteHeader(http.StatusOK)
	timeline.SaveTimelinesAndPosts(timelines, username, *identityFilePath)

}

func unfollowHandler(reply http.ResponseWriter, request *http.Request) {
	if !loggedIn {
		bytes, _ := json.Marshal("Not logged in")
		reply.WriteHeader(http.StatusUnauthorized)
		reply.Write(bytes)
		return
	}

	query := request.URL.Query()

	timeline.UnfollowUser(&timelines, query.Get("usernameToUnfollow"))

	reply.WriteHeader(http.StatusOK)
	timeline.SaveTimelinesAndPosts(timelines, username, *identityFilePath)

}

func updateHandler(reply http.ResponseWriter, request *http.Request) {
	if !loggedIn {
		bytes, _ := json.Marshal("Not logged in")
		reply.WriteHeader(http.StatusUnauthorized)
		reply.Write(bytes)
		return
	}

	// On message from pubsub topic, ask dht for providers of the post cid -> Get it and annouce ourselves as providers of it
	timeline.UpdateTimeline(timelines) // Gets all the pending posts for each subscribed timeline

	for _, timeline := range timelines {
		retrievedCIDS := []*cid.Cid{}
		for _, postCid := range timeline.PendingPosts {
			addr, _ := peerDHT.FindProviders(*postCid)
			fmt.Println(addr)

			for _, peer := range addr {
				post, err := postretrieval.RetrievePost(ctx, peerHost, peer, *postCid)
				if err != nil {
					fmt.Println(err)
					continue
				}
				retrievedCIDS = append(retrievedCIDS, postCid)
				fmt.Println(post)
				timeline.Posts[postCid.String()] = *post
				contentRouting.ProvideNewPost(postCid, peerDHT, timeline.Owner)
				break
			}
		}
		newPendingPosts := []*cid.Cid{}
		for _, cid := range timeline.PendingPosts {
			if !slices.Contains(retrievedCIDS, cid) {
				newPendingPosts = append(newPendingPosts, cid)
			}
		}
		timeline.PendingPosts = newPendingPosts
	}

	reply.WriteHeader(http.StatusOK)
	timeline.SaveTimelinesAndPosts(timelines, username, *identityFilePath)
}

func allTimelineHandler(reply http.ResponseWriter, request *http.Request) {
	jsonTimelines, err := timeline.TimelinesToJSON(timelines)

	if err != nil {
		bytes, _ := json.Marshal("Error converting timelines to json")
		reply.WriteHeader(http.StatusInternalServerError)
		reply.Write(bytes)
		return
	}

	reply.Write(jsonTimelines)
}

func timelineHandler(reply http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()

	jsonTimeline, err := timeline.TimelineToJSON(timelines, query.Get("username"))
	if err != nil {
		bytes, _ := json.Marshal("Error converting timeline to json")
		reply.WriteHeader(http.StatusInternalServerError)
		reply.Write(bytes)
		return
	}

	reply.Write(jsonTimeline)
}
