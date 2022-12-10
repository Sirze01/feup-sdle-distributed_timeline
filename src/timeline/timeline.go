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

// ChatRoomBufSize is the number of incoming messages to buffer for each topic.
const ChatRoomBufSize = 128

var timelineLogger = log.Logger("rettiwt-timeline")

// ChatRoom represents a subscription to a single PubSub topic. Messages
// can be published to the topic with ChatRoom.Publish, and received
// messages are pushed to the Messages channel.
type ChatRoom struct {
	// Messages is a channel of messages received from other peers in the chat room
	Messages []*ChatMessage

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	roomName string
	self     peer.ID
	nick     string
}

// ChatMessage gets converted to/from JSON and sent in the body of pubsub messages.
type ChatMessage struct {
	Message    string
	SenderNick string
	TimeStamp  time.Time
}

// JoinChatRoom tries to subscribe to the PubSub topic for the room name, returning
// a ChatRoom on success.
func JoinChatRoom(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, nickname string, roomName string) (*ChatRoom, error) {
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

	cr := &ChatRoom{
		ctx:      ctx,
		ps:       ps,
		topic:    topic,
		sub:      sub,
		self:     selfID,
		nick:     nickname,
		roomName: roomName,
		Messages: []*ChatMessage{},
	}

	// start reading messages from the subscription in a loop

	return cr, nil
}

// Publish sends a message to the pubsub topic.
func (cr *ChatRoom) Publish(message string) error {
	m := ChatMessage{
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

func (cr *ChatRoom) ListPeers() []peer.ID {
	return cr.ps.ListPeers(cr.roomName)
}

func GetFollowers(timelines []*ChatRoom, dht *dht.KademliaDHT, roomname string) []string {
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

func FollowUser(timelines []*ChatRoom, ps *pubsub.PubSub, ctx context.Context, selfID peer.ID, nickname string, roomName string) []*ChatRoom {
	timeline, err := JoinChatRoom(ctx, ps, selfID, nickname, roomName)
	if err != nil {
		fmt.Println("Error joining chat room: ", err)
		return timelines
	}
	timelines = append(timelines, timeline)
	return timelines
}

func UnfollowUser(timelines []*ChatRoom, roomName string) []*ChatRoom {
	var newTimelines []*ChatRoom
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

func (cr *ChatRoom) readLoop(c chan struct{}) {
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
		cm := new(ChatMessage)
		err = json.Unmarshal(msg.Data, cm)
		if err != nil {
			continue
		}
		// send valid messages onto the Messages channel

		cr.Messages = append(cr.Messages, cm)

	}
}

func updateUserTimeline(userTimeline *ChatRoom, wg *sync.WaitGroup) {
	c := make(chan struct{})
	go userTimeline.readLoop(c)
	select {
	case <-c:

	case <-time.After(1 * time.Second):

	}
	wg.Done()
}

func UpdateTimeline(timelines []*ChatRoom) {
	allMessages := []*ChatMessage{}

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

func StartTimelines(username string, ps *pubsub.PubSub, ctx context.Context, selfID peer.ID) ([]*ChatRoom, *ChatRoom) {
	var timelines []*ChatRoom
	var ownTimeline *ChatRoom
	var timeline_json_file []byte
	timeline_json := make(map[string][]ChatMessage)
	if _, err := os.Stat("./nodes/" + username + ".timelines.json"); err != nil {

		generalTimeline, err := JoinChatRoom(ctx, ps, selfID, username, "rettiwt")
		if err != nil {
			panic(err)
		}
		ownTimeline, err := JoinChatRoom(ctx, ps, selfID, username, username)
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

		msgs_pointers := []*ChatMessage{}
		for _, msg := range messages {
			fmt.Print("\n\n\n\n")
			fmt.Println(msg)
			msg_pointer := &ChatMessage{}
			*msg_pointer = msg

			msgs_pointers = append(msgs_pointers, msg_pointer)
		}

		cr := &ChatRoom{
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

func DownloadTimelines(timelines []*ChatRoom, username string) {
	timeline_json := make(map[string][]ChatMessage)

	for _, timeline := range timelines {
		dereferenced_msgs := []ChatMessage{}
		for _, msg := range timeline.Messages {
			dereferenced_msgs = append(dereferenced_msgs, *msg)
		}
		timeline_json[timeline.roomName] = dereferenced_msgs
	}

	json, err := json.Marshal(timeline_json)
	if err != nil {
		fmt.Println("Error marshalling timeline json: ", err)
	}
	err = os.WriteFile("./nodes/"+username+".timelines.json", json, 0666)
	if err != nil {
		fmt.Println("Error writing timeline json: ", err)
	}
}
