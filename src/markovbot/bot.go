package markovbot

import (
	"fmt"
	"github.com/nlopes/slack"
	"github.com/syndtr/goleveldb/leveldb"
	"log"
	"math"
	"time"
)

type MarkovBot interface {
	Start()
}

type markovBot struct {
	db               *leveldb.DB
	slack            SlackApi
	markovOrder      int
	maxResponseWords int
	maxAgeHours      int

	// single thread of control for backfilling channels by channelId
	channelBackfill chan backfillEvent
}

type backfillEvent struct {
	channelId  string
	fromUserId string
}

func NewMarkovBot(
	apiToken,
	dbFilename string,
	markovOrder,
	maxResponseWords int,
	maxAgeHours int,
) (MarkovBot, error) {
	db, err := leveldb.OpenFile(dbFilename, nil)
	if err != nil {
		return nil, err
	}

	b := &markovBot{
		db:               db,
		slack:            NewSlackApi(apiToken),
		markovOrder:      markovOrder,
		maxResponseWords: maxResponseWords,
		maxAgeHours:      maxAgeHours,
		channelBackfill:  make(chan backfillEvent, 1024),
	}

	// TODO start background for pruning?
	for i := 0; i < 3; i++ {
		go b.startBackfillLoop()
	}

	return b, nil
}

func (b *markovBot) startBackfillLoop() {
	for {
		select {
		case evt := <-b.channelBackfill:
			channelId, ignoreUserId := evt.channelId, evt.fromUserId
			channelName, err := b.slack.GetChannelName(channelId)
			if err != nil {
				channelName = channelId
			}

			startTime := time.Now()
			lastStoredTimestamp, err := b.lastTimestampForChannel(channelId)
			if err != nil {
				log.Println("Error fetching last timestamp for", channelName, err)
				continue
			}
			// only keep data up to the requested limit
			lastTimestamp := math.Max(
				lastStoredTimestamp,
				float64(time.Now().Add(time.Hour*time.Duration(-b.maxAgeHours)).Unix()),
			)

			msgs, err := b.slack.GetMessages(channelId, lastTimestamp, ignoreUserId)
			if err != nil {
				log.Printf("Couldn't backfill channel %v: %v\n", channelId, err)
				return
			}
			for _, msg := range msgs {
				b.addMessage(msg.text, channelId, msg.userId, msg.timestamp)
			}
			log.Printf("Finished backfill for %s: %d messages in %v", channelName, len(msgs), time.Since(startTime))
		}
	}
}

func (b *markovBot) queueChannelBackfill(channelId, fromUserId string) {
	log.Println("Queueing backfill for", channelId)
	b.channelBackfill <- backfillEvent{
		channelId:  channelId,
		fromUserId: fromUserId,
	}
}

func (b *markovBot) handleLoop(rtm *slack.RTM) {
Loop:
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			switch msg.Data.(type) {
			case *slack.InvalidAuthEvent:
				fmt.Printf("Invalid credentials")
				break Loop

			case *slack.ConnectedEvent:
				connectedEvent := msg.Data.(*slack.ConnectedEvent)
				for _, channel := range connectedEvent.Info.Channels {
					if channel.IsMember {
						b.queueChannelBackfill(channel.ID, connectedEvent.Info.User.ID)
					}
				}
				// includes all groups user is a member of
				for _, group := range connectedEvent.Info.Groups {
					b.queueChannelBackfill(group.ID, connectedEvent.Info.User.ID)
				}

			case *slack.GroupJoinedEvent:
				// joined a new channel
				joinedEvent := msg.Data.(*slack.GroupJoinedEvent)
				b.queueChannelBackfill(joinedEvent.Channel.ID, rtm.GetInfo().User.ID)

			case *slack.ChannelJoinedEvent:
				// joined a new channel
				joinedEvent := msg.Data.(*slack.ChannelJoinedEvent)
				b.queueChannelBackfill(joinedEvent.Channel.ID, rtm.GetInfo().User.ID)

			case *slack.MessageEvent:
				message := msg.Data.(*slack.MessageEvent)
				if rtm.GetInfo().User.ID == message.User {
					// skip messages that the bot has sent
					continue
				}

				if seed := getMessageSeed(message.Text, message.Channel); seed != nil {
					rtm.SendMessage(rtm.NewTypingMessage(message.Channel))

					log.Printf("Responding to: %v in %v\n", message.Text, message.Channel)
					response, err := b.generateResponse(seed)
					if err != nil {
						log.Printf("Error while generating response for %v: %v\n", message.Text, err)
					}

					rtm.SendMessage(rtm.NewOutgoingMessage(response, message.Channel))
				} else if message.Type == "message" && message.SubType == "" {
					b.addMessage(message.Text, message.Channel, message.User, message.Timestamp)
					log.Printf("Adding: %v '%v' in %v\n", message.User, message.Text, message.Channel)
				}

			default:
				// Ignore everything else
			}
		}
	}
}

func (b *markovBot) Start() {
	rtm := b.slack.GetClient().NewRTM()

	// kick off the socket and handle all the responses!
	go rtm.ManageConnection()
	b.handleLoop(rtm)
}
