package timeline

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht"
	recordpeer "git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht/record/rettiwt-peer"
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
	// Messages is a channel of messages received from other peers in the chat room
	Messages []*TimelineMessage

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	roomName string
	self     peer.ID
	nick     string
}

// TimelineMessage gets converted to/from JSON and sent in the body of pubsub messages.
type TimelineMessage struct {
	Message    string
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
		ctx:      ctx,
		ps:       ps,
		topic:    topic,
		sub:      sub,
		self:     selfID,
		nick:     nickname,
		roomName: roomName,
		Messages: []*TimelineMessage{},
	}

	// start reading messages from the subscription in a loop

	return cr, nil
}

// Publish sends a message to the pubsub topic.
func (cr *UserTimeline) Publish(message string) error {
	m := TimelineMessage{
		Message:    message,
		SenderNick: cr.nick,
		TimeStamp:  time.Now(),
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}

	cr.Messages = append(cr.Messages, &m)

	return cr.topic.Publish(cr.ctx, msgBytes)
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
		cm := new(TimelineMessage)
		err = json.Unmarshal(msg.Data, cm)
		if err != nil {
			continue
		}
		// send valid messages onto the Messages channel

		cr.Messages = append(cr.Messages, cm)

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
	allMessages := []*TimelineMessage{}

	fmt.Println("Updating timeline...")
	wg := sync.WaitGroup{}
	for _, timeline := range timelines {
		wg.Add(1)
		go updateUserTimeline(timeline, &wg)
	}
	wg.Wait()

	for _, timeline := range timelines {
		allMessages = append(allMessages, timeline.Messages...)
	}

	sort.Slice(allMessages, func(i, j int) bool {
		return allMessages[i].TimeStamp.After(allMessages[j].TimeStamp)
	})

	for _, message := range allMessages {
		fmt.Println("\n\nFrom: ", message.SenderNick, "\nMessage: ", message.Message, "\nTime: ", message.TimeStamp.Format("2006-01-02 15:04:05"))

	}

}

func StartTimelines(username string, ps *pubsub.PubSub, ctx context.Context, selfID peer.ID) ([]*UserTimeline, *UserTimeline) {
	var timelines []*UserTimeline
	var ownTimeline *UserTimeline
	var timeline_json_file []byte
	timeline_json := make(map[string][]TimelineMessage)
	if _, err := os.Stat("./nodes/" + username + ".timelines.json"); err != nil {

		generalTimeline, err := JoinUserTimeline(ctx, ps, selfID, username, "rettiwt")
		if err != nil {
			panic(err)
		}
		ownTimeline, err := JoinUserTimeline(ctx, ps, selfID, username, username)
		if err != nil {
			panic(err)
		}
		timelines = append(timelines, ownTimeline)
		timelines = append(timelines, generalTimeline)

		return timelines, ownTimeline

	}

	timeline_json_file, err := os.ReadFile("./nodes/" + username + ".timelines.json")
	if err != nil {
		fmt.Println("Error reading timeline json: ", err)
	}
	err = json.Unmarshal(timeline_json_file, &timeline_json)
	if err != nil {
		fmt.Println("Error unmarshalling timeline json: ", err)
	}

	for user, messages := range timeline_json {
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

		msgs_pointers := []*TimelineMessage{}
		for _, msg := range messages {
			msg_pointer := &TimelineMessage{}
			*msg_pointer = msg

			msgs_pointers = append(msgs_pointers, msg_pointer)
		}

		cr := &UserTimeline{
			ctx:      ctx,
			ps:       ps,
			topic:    topic,
			sub:      sub,
			self:     selfID,
			nick:     username,
			roomName: user,
			Messages: msgs_pointers,
		}

		if user == username {
			ownTimeline = cr
		}

		timelines = append(timelines, cr)
	}

	return timelines, ownTimeline

}

func DownloadTimelines(timelines []*UserTimeline, username string) {
	timeline_json := make(map[string][]TimelineMessage)

	for _, timeline := range timelines {
		dereferenced_msgs := []TimelineMessage{}
		for _, msg := range timeline.Messages {
			dereferenced_msgs = append(dereferenced_msgs, *msg)
		}
		timeline_json[timeline.roomName] = dereferenced_msgs
	}

	json, err := json.MarshalIndent(timeline_json, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling timeline json: ", err)
	}
	err = os.WriteFile("./nodes/"+username+".timelines.json", json, 0666)
	if err != nil {
		fmt.Println("Error writing timeline json: ", err)
	}
}
