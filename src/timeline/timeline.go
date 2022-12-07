package timeline

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"sync"

	log "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/peer"

	"time"

	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

var timelineLogger = log.Logger("rettiwt-timeline")

// ChatRoomBufSize is the number of incoming messages to buffer for each topic.
const TimeLineBufSize = 128

// ChatRoom represents a subscription to a single PubSub topic. Messages
// can be published to the topic with ChatRoom.Publish, and received
// messages are pushed to the Messages channel.

var TimeLines = []*UserTimeLine{}

type UserTimeLine struct {
	// Messages is a channel of messages received from other peers in the chat room
	Messages chan *Message

	ctx   context.Context
	ps    *pubsub.PubSub
	topic *pubsub.Topic
	sub   *pubsub.Subscription

	self         peer.ID
	followedUser string
}

// ChatMessage gets converted to/from JSON and sent in the body of pubsub messages.
type Message struct {
	Message    string
	SenderNick string
	TimeStamp  time.Time
}

// JoinChatRoom tries to subscribe to the PubSub topic for the room name, returning
// a ChatRoom on success.
func FollowUser(ctx context.Context, ps *pubsub.PubSub, selfID peer.ID, userName string) (*UserTimeLine, error) {

	//check if the user exists

	//TODO

	//check if the user is already followed
	for _, timeline := range TimeLines {
		if timeline.followedUser == userName {
			return timeline, nil
		}
	}

	// join the pubsub topic
	topic, err := ps.Join(topicName(userName))
	if err != nil {
		return nil, err
	}

	// and subscribe to it
	sub, err := topic.Subscribe()
	if err != nil {
		return nil, err
	}

	cr := &UserTimeLine{
		ctx:          ctx,
		ps:           ps,
		topic:        topic,
		sub:          sub,
		followedUser: userName,
		self:         selfID,
		Messages:     make(chan *Message, TimeLineBufSize),
	}

	TimeLines = append(TimeLines, cr)

	// start reading messages from the subscription in a loop
	//go cr.readLoop()
	return cr, nil
}

func UnfollowUser(ctx context.Context, ps *pubsub.PubSub, userName string) error {
	var idxRemove int = -1
	for index, timeline := range TimeLines {
		if timeline.followedUser == userName {
			//unsubscribe from the pubsub topic
			timeline.sub.Cancel()
			idxRemove = index

		}
	}
	//delete the timeline from the list of timeline
	if idxRemove != -1 {
		copy(TimeLines[idxRemove:], TimeLines[idxRemove+1:])
		TimeLines[len(TimeLines)-1] = nil
		TimeLines = TimeLines[:len(TimeLines)-1]
		return nil
	}

	return fmt.Errorf("user not found")

}

// Publish sends a message to the pubsub topic.
func Publish(message string, username string) error {
	m := Message{
		Message:    message,
		SenderNick: username,
		TimeStamp:  time.Now(),
	}
	msgBytes, err := json.Marshal(m)
	if err != nil {
		return err
	}
	for _, timeline := range TimeLines {
		if timeline.followedUser == username {
			return timeline.topic.Publish(timeline.ctx, msgBytes)
		}
	}

	return fmt.Errorf("user not found")
}

func (cr *UserTimeLine) ListPeers() []peer.ID {
	return cr.ps.ListPeers(topicName(cr.followedUser))
}

// readLoop pulls messages from the pubsub topic and pushes them onto the Messages channel.
func (cr *UserTimeLine) readLoop(c chan struct{}) {
	defer close(c)
	for {

		msg, err := cr.sub.Next(cr.ctx)
		if err != nil {
			close(cr.Messages)
			return
		}
		// only forward messages delivered by others
		if msg.ReceivedFrom == cr.self {
			continue
		}
		cm := new(Message)
		err = json.Unmarshal(msg.Data, cm)
		if err != nil {
			continue
		}
		// send valid messages onto the Messages channel
		cr.Messages <- cm

	}
}

func ChanToSlice(ch interface{}) interface{} {
	chv := reflect.ValueOf(ch)
	slv := reflect.MakeSlice(reflect.SliceOf(reflect.TypeOf(ch).Elem()), 0, 0)
	for {
		v, ok := chv.Recv()
		if !ok {
			return slv.Interface()
		}
		slv = reflect.Append(slv, v)
	}
}

func updateUserTimeline(userTimeline *UserTimeLine, wg *sync.WaitGroup) {
	c := make(chan struct{})
	go userTimeline.readLoop(c)
	select {
	case <-c:
		fmt.Println("Finished readloop (impossible)")
	case <-time.After(1 * time.Second):
		fmt.Println("Finished readloop (timeout)")
	}
	wg.Done()
}

func UpdateTimeline() {
	allMessages := []Message{}

	fmt.Println("Updating timeline")
	wg := sync.WaitGroup{}
	for _, timeline := range TimeLines {
		wg.Add(1)
		go updateUserTimeline(timeline, &wg)
	}
	wg.Wait()

	fmt.Println("All timelines are updated")
	for _, timeline := range TimeLines {
		timeline_msgs := ChanToSlice(timeline.Messages).([]Message)
		fmt.Println("Timeline messages are collected")
		allMessages = append(allMessages, timeline_msgs...)
		fmt.Println("Timeline messages are appended")
	}

	fmt.Println("All messages are collected")

	sort.Slice(allMessages, func(i, j int) bool {
		return allMessages[i].TimeStamp.After(allMessages[j].TimeStamp)
	})

	fmt.Println("All messages are sorted")

	for _, message := range allMessages {
		fmt.Println(message)
	}

}

func topicName(username string) string {
	return "username:" + username
}
