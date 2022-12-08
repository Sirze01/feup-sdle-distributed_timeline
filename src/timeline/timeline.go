package timeline

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	log "github.com/ipfs/go-log/v2"

	"git.fe.up.pt/sdle/2022/t3/g15/proj2/proj2/core/dht"
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
	SenderID   string
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
		SenderID:   cr.self.Pretty(),
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
				username, err := dht.GetValue(peer.String())
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
