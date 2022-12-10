package timeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht"
	recordpeer "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht/record/rettiwt-peer"
	"github.com/ipfs/go-cid"
	log "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p/core/peer"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

// UserTimelineBufSize is the number of incoming messages to buffer for each topic.
const UserTimelineBufSize = 128

var timelineLogger = log.Logger("rettiwt-timeline")

// UserTimeline represents a subscription to a single PubSub topic. Messages
// can be published to the topic with UserTimeline.Publish, and received
// messages are pushed to the Messages channel.
type UserTimeline struct {
	// Posts is a channel of messages received from other peers in the chat room
	Posts        map[string]TimelinePost
	PendingPosts []*cid.Cid

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	roomName   string
	self       peer.ID
	nick       string
	CurrPostID int
}

// TimelinePost gets converted to/from JSON and sent in the body of pubsub messages.
type TimelinePost struct {
	Content    string
	SenderNick string
	TimeStamp  time.Time
}

// JoinUserTimeline tries to subscribe to the PubSub topic for the room name, returning
// a UserTimeline on success.
func JoinUserTimeline(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, nickname string, roomName string) (*UserTimeline, error) {
	// join the pubsub topic
	topic, err := ps.Join(roomName)
	if err != nil {
		return nil, err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	cr := &UserTimeline{
		ctx:        ctx,
		ps:         ps,
		topic:      topic,
		sub:        sub,
		self:       selfID,
		nick:       nickname,
		roomName:   roomName,
		Posts:      make(map[string]TimelinePost),
		CurrPostID: -1,
	}

	// start reading messages from the subscription in a loop

	return cr, nil
}

// Publish sends a message to the pubsub topic.
func (cr *UserTimeline) NewPost(cid *cid.Cid, content string) TimelinePost {
	m := TimelinePost{
		Content:    content,
		SenderNick: cr.nick,
		TimeStamp:  time.Now(),
	}

	cr.Posts[cid.String()] = m
	return m
}

func (cr *UserTimeline) Publish(announcement string) error {
	return cr.topic.Publish(cr.ctx, []byte(announcement))
}

func (cr *UserTimeline) ListPeers() []peer.ID {
	return cr.ps.ListPeers(cr.roomName)
}

func GetFollowers(timelines []*UserTimeline, dht *dht.KademliaDHT, roomname string) []string {
	var users = []string{}
	for _, timeline := range timelines {
		if timeline.roomName == roomname {
			for _, peer := range timeline.ListPeers() {
				fmt.Println("Peer: ", peer.String())
				username, err := dht.GetValue("/" + recordpeer.RettiwtPeerNS + "/" + peer.String())
				if err != nil {
					users = append(users, string(username))
				} else {
					timelineLogger.Error(err)
				}
			}
		}
	}

	fmt.Println("Users: ", users)
	return users

}

func FollowUser(timelines []*UserTimeline, ps *pubsub.PubSub, ctx context.Context, selfID peer.ID, nickname string, roomName string) []*UserTimeline {
	timeline, err := JoinUserTimeline(ctx, ps, selfID, nickname, roomName)
	if err != nil {
		fmt.Println("Error joining chat room: ", err)
		return timelines
	}
	timelines = append(timelines, timeline)
	return timelines
}

func UnfollowUser(timelines []*UserTimeline, roomName string) []*UserTimeline {
	var newTimelines []*UserTimeline
	for _, timeline := range timelines {
		if timeline.roomName != roomName {
			newTimelines = append(newTimelines, timeline)
		} else {
			timeline.sub.Cancel()
		}
	}
	return newTimelines
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.

func (cr *UserTimeline) readLoop(c chan struct{}) {
	defer close(c)
	for {

		msg, err := cr.sub.Next(cr.ctx)
		if err != nil {
			return
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom == cr.self {
			continue
		}

		var pendingPostCid cid.Cid
		err = pendingPostCid.UnmarshalJSON(msg.Data)
		if err != nil {
			return
		}

		cr.PendingPosts = append(cr.PendingPosts, &pendingPostCid)
	}
}

func updateUserTimeline(userTimeline *UserTimeline, wg *sync.WaitGroup) {
	c := make(chan struct{})
	go userTimeline.readLoop(c)
	select {
	case <-c:

	case <-time.After(1 * time.Second):

	}
	wg.Done()
}

func UpdateTimeline(timelines []*UserTimeline) {
	allPendingPosts := []*cid.Cid{}

	fmt.Println("Updating timeline...")
	wg := sync.WaitGroup{}
	for _, timeline := range timelines {
		wg.Add(1)
		go updateUserTimeline(timeline, &wg)
	}
	wg.Wait()

	for _, timeline := range timelines {
		allPendingPosts = append(allPendingPosts, timeline.PendingPosts...)
	}

	// TODO: Log instead of printing
	fmt.Println("CIDS:")
	for _, cid := range allPendingPosts {
		fmt.Println(cid.String())
	}

}

func StartTimelines(username string, ps *pubsub.PubSub, ctx context.Context, selfID peer.ID, postStoragePath string) ([]*UserTimeline, *UserTimeline) {
	var timelines []*UserTimeline
	var ownTimeline *UserTimeline

	if _, err := os.Stat(filepath.Dir(postStoragePath) + "/" + username + ".timelines.json"); err != nil {

		generalTimeline, err := JoinUserTimeline(ctx, ps, selfID, username, "rettiwt")
		if err != nil {
			panic(err)
		}
		ownTimeline, err = JoinUserTimeline(ctx, ps, selfID, username, username)
		if err != nil {
			panic(err)
		}
		timelines = append(timelines, ownTimeline)
		timelines = append(timelines, generalTimeline)

		return timelines, ownTimeline

	}

	timelinesJSONFile, err := os.ReadFile(filepath.Dir(postStoragePath) + "/" + username + ".timelines.json")
	if err != nil {
		fmt.Println("Error reading timeline json: ", err)
	}

	var timelinesJSON = map[string]map[string]TimelinePost{}

	err = json.Unmarshal(timelinesJSONFile, &timelinesJSON)
	if err != nil {
		fmt.Println("Error unmarshalling timeline json: ", err)
	}

	for user, posts := range timelinesJSON {
		topic, err := ps.Join(user)
		if err != nil {
			fmt.Println("Error joining topic: ", err)
			continue

		}

		// and subscribe to it
		sub, err := topic.Subscribe()
		if err != nil {
			fmt.Println("Error subscribing topic: ", err)
			continue

		}

		cr := &UserTimeline{
			ctx:        ctx,
			ps:         ps,
			topic:      topic,
			sub:        sub,
			self:       selfID,
			nick:       username,
			roomName:   user,
			Posts:      posts,
			CurrPostID: len(posts) - 1, // TODO: Check me
		}

		if user == username {
			ownTimeline = cr
		}

		timelines = append(timelines, cr)
	}

	return timelines, ownTimeline

}

func SaveTimelinesAndPosts(timelines []*UserTimeline, username, postStoragePath string) {
	var timelinesToJSON = map[string]map[string]TimelinePost{}

	for _, timeline := range timelines {
		timelinesToJSON[timeline.roomName] = timeline.Posts
	}

	json, err := json.MarshalIndent(timelinesToJSON, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling timeline json: ", err)
	}
	err = os.WriteFile(filepath.Dir(postStoragePath)+"/"+username+".timelines.json", json, 0666)
	if err != nil {
		fmt.Println("Error writing timeline json: ", err)
	}
}
