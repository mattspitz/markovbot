package markovbot

import (
	"errors"
	"github.com/nlopes/slack"
	cache "github.com/robfig/go-cache"
	"strconv"
	"strings"
	"time"
)

const (
	userExpirationDuration    = time.Hour
	userCleanupInterval       = time.Minute * 30
	channelExpirationDuration = time.Minute * 10
	channelCleanupInterval    = time.Minute * 30
)

type SlackApi interface {
	GetChannelName(channelId string) (string, error)
	GetClient() *slack.Client
	GetMessages(channelId string, sinceTimestamp float64, ignoreUserId string) ([]*Message, error)
	GetUserFromUsername(username string) (string, error)
}

type Message struct {
	text      string
	userId    string
	timestamp string
}

type slackApi struct {
	apiToken     string
	userCache    *cache.Cache
	channelCache *cache.Cache
}

func NewSlackApi(apiToken string) SlackApi {
	return &slackApi{
		apiToken:     apiToken,
		channelCache: cache.New(channelExpirationDuration, channelCleanupInterval),
		userCache:    cache.New(userExpirationDuration, userCleanupInterval),
	}
}

func (s *slackApi) GetClient() *slack.Client {
	return slack.New(s.apiToken)
}

func (s *slackApi) GetMessages(channelId string, sinceTimestamp float64, ignoreUserId string) ([]*Message, error) {
	var messages []*Message
	oldest := strconv.FormatFloat(sinceTimestamp, 'f', -1, 64)
	for {
		var f func(string, slack.HistoryParameters) (*slack.History, error)
		if strings.HasPrefix(channelId, "C") {
			f = s.GetClient().GetChannelHistory
		} else if strings.HasPrefix(channelId, "G") {
			f = s.GetClient().GetGroupHistory
		}
		history, err := f(
			channelId, slack.HistoryParameters{
				Oldest:    oldest,
				Count:     1000,
				Inclusive: false,
			})
		if err != nil {
			return nil, err
		}
		for _, msg := range history.Messages {
			// TODO unify the message filter logic (it's also in the MessageEvent handler)

			// ignore messages from self
			if msg.User == ignoreUserId {
				continue
			}

			// if this would trigger a message, ignore it
			if getMessageSeed(msg.Text, msg.Channel) != nil {
				continue
			}

			if msg.Type == "message" && msg.SubType == "" {
				messages = append(messages, &Message{
					text:      msg.Text,
					userId:    msg.User,
					timestamp: msg.Timestamp,
				})
				// next time, fetch newer things
				if msg.Timestamp > oldest {
					oldest = msg.Timestamp
				}
			}
		}
		if !history.HasMore {
			return messages, nil
		}
	}
}

func (s *slackApi) GetChannelName(channelId string) (string, error) {
	if _, ok := s.channelCache.Get("channels"); !ok {
		channels, err := s.GetClient().GetChannels(false)
		if err != nil {
			return "", err
		}
		groups, err := s.GetClient().GetGroups(false)
		if err != nil {
			return "", err
		}

		channelMap := make(map[string]string)
		for _, channel := range channels {
			channelMap[channel.ID] = channel.Name
		}
		for _, group := range groups {
			channelMap[group.ID] = group.Name
		}

		s.channelCache.Add("channels", channelMap, channelExpirationDuration)
	}

	val, ok := s.channelCache.Get("channels")
	if !ok {
		return "", errors.New("Didn't we just put a channel map in here?")
	}

	channelName := val.(map[string]string)[channelId]
	if channelName != "" {
		return "#" + channelName, nil
	} else {
		return channelId, nil
	}
}

func (s *slackApi) GetUserFromUsername(username string) (string, error) {
	if _, ok := s.userCache.Get("users"); !ok {
		users, err := s.GetClient().GetUsers()
		if err != nil {
			return "", err
		}
		userMap := make(map[string]string)
		for _, user := range users {
			userMap[strings.ToLower(user.Name)] = user.ID
		}
		s.userCache.Add("users", userMap, userExpirationDuration)
	}

	val, ok := s.userCache.Get("users")
	if !ok {
		return "", errors.New("Didn't we just put a user map in here?")
	}

	return val.(map[string]string)[strings.ToLower(username)], nil
}
