package markovbot

import (
	"fmt"
	"markovbot/markov"
	"regexp"
	"strings"
)

type messageSeed struct {
	targets []string
	channel string
}

var messageRegex = regexp.MustCompile(".*markovbot:?\\s*(.*)")
var targetRegex = regexp.MustCompile("\\b\\w+\\b")

/* Returns the seed for a message, either a list of users or a channel name */
func getMessageSeed(text, channel string) *messageSeed {
	if msgGroups := messageRegex.FindStringSubmatch(text); msgGroups != nil {
		target := msgGroups[1]
		if targetGroups := targetRegex.FindAllString(target, -1); targetGroups != nil {
			// attempt to pull target usernames first
			return &messageSeed{
				targets: targetGroups,
				channel: channel,
			}
		} else {
			// no particular targets, fall back on channel
			return &messageSeed{
				channel: channel,
			}
		}
	}
	return nil
}

func (b *markovBot) generateResponse(seed *messageSeed) (string, error) {
	var sources []string
	var messages []string
	if seed.targets != nil {
		// generate from users
		for _, targetName := range seed.targets {
			userId, err := b.slack.GetUserFromUsername(targetName)
			if err != nil {
				return "", err
			}
			if userId != "" {
				userMessages, err := b.getMessagesForChannelUser(seed.channel, userId)
				if err != nil {
					return "", err
				}
				messages = append(messages, userMessages...)
				sources = append(sources, targetName)
			}
		}
	}
	if len(messages) == 0 {
		channelName, err := b.slack.GetChannelName(seed.channel)
		if err != nil {
			return "", err
		}

		// if no targets or malformed targets, try the channel
		channelMessages, err := b.getMessagesForChannel(seed.channel)
		if err != nil {
			return "", err
		}
		messages = channelMessages
		sources = append(sources, channelName)
	}

	model := markov.NewMarkovModel(b.markovOrder)
	for _, msg := range messages {
		model.AddMessage(msg)
	}

	sentence := model.GenerateSentence(b.maxResponseWords)
	response := fmt.Sprintf(
		"\"%s\" - %s",
		sentence, strings.Join(sources, "/"),
	)
	return response, nil
}
