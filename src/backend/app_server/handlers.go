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
)

func registerHandler(reply http.ResponseWriter, request *http.Request) {
	fmt.Println("Registering user")
	query := request.URL.Query()

	if query.Get("username") == "" || query.Get("password") == "" {
		bytes, _ := json.Marshal("Invalid username or password")
		reply.WriteHeader(http.StatusBadRequest)
		reply.Write(bytes)
		return
	}

	err := peer.RegisterUser(true, dht, query.Get("username"), query.Get("password"))
	if err != nil {
		bytes, _ := json.Marshal(err)
		reply.WriteHeader(http.StatusBadRequest)
		reply.Write(bytes)
		return
	}

	reply.WriteHeader(http.StatusOK)
}

func loginHandler(reply http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()

	if query.Get("username") == "" || query.Get("password") == "" {
		bytes, _ := json.Marshal("Invalid username or password")
		reply.WriteHeader(http.StatusBadRequest)
		reply.Write(bytes)
		return
	}

	err := peer.LoginUser(dht, query.Get("username"), query.Get("password"))
	if err != nil {
		bytes, _ := json.Marshal(err)
		reply.WriteHeader(http.StatusBadRequest)
		reply.Write(bytes)
		return
	}

	reply.WriteHeader(http.StatusOK)
}

func publishHandler(reply http.ResponseWriter, request *http.Request) {
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

	a := *host
	cid := contentRouting.NewCID(personalTimeline, a.ID().String())
	personalTimeline.NewPost(cid, post)
	contentRouting.ProvideNewPost(cid, dht, personalTimeline.Owner)
	contentRouting.AnounceNewPost(personalTimeline, *cid)

	reply.WriteHeader(http.StatusOK)
	timeline.SaveTimelinesAndPosts(timelines, personalTimeline.Owner, *identityFilePath)

}

func followHandler(reply http.ResponseWriter, request *http.Request) {

	query := request.URL.Query()

	if query.Get("usernameToFollow") == personalTimeline.Owner {
		bytes, _ := json.Marshal("Can't follow yourself")
		reply.WriteHeader(http.StatusBadRequest)
		reply.Write(bytes)
		return
	}

	// Ask dht for history
	a := *host
	_, currTimeline := timeline.FollowUser(&timelines, globalpubsub, *globalContext, a.ID(), query.Get("usernameToFollow"))

	marshaledPeerRecord, _ := dht.GetValue("/" + peerns.RettiwtPeerNS + "/" + currTimeline.Owner) // TODO: Handle error

	peerRecord := contentRouting.PeerRecordUnmarshalJson(marshaledPeerRecord)
	for _, cidRecord := range peerRecord.CidsCache {
		if !cidRecord.ExpireDate.After(time.Now()) {
			continue
		}

		addr, _ := dht.FindProviders(cidRecord.Cid)
		fmt.Println(addr)

		for _, peer := range addr {
			post, err := postretrieval.RetrievePost(*globalContext, *host, peer, cidRecord.Cid)
			if err != nil {
				fmt.Println(err)
				continue
			}
			fmt.Println(post)
			currTimeline.Posts[cidRecord.Cid.String()] = *post
			contentRouting.ProvideNewPost(&cidRecord.Cid, dht, currTimeline.Owner)
			break
		}
	}
	// Ask dht for providers for each post cid -> Get them and annouce ourselves as providers of them
	// Follow the user pubsub topic

	reply.WriteHeader(http.StatusOK)
	timeline.SaveTimelinesAndPosts(timelines, personalTimeline.Owner, *identityFilePath)

}

func unfollowHandler(reply http.ResponseWriter, request *http.Request) {
	query := request.URL.Query()

	timeline.UnfollowUser(&timelines, query.Get("usernameToUnfollow"))

	reply.WriteHeader(http.StatusOK)
	timeline.SaveTimelinesAndPosts(timelines, personalTimeline.Owner, *identityFilePath)

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
