package markovbot

import (
	"fmt"
	leveldb_util "github.com/syndtr/goleveldb/leveldb/util"
	"log"
	"math"
	"strconv"
	"strings"
)

// TODO func startCleanupLoop(interval)

func makeDbKey(channel, user, timestamp string) []byte {
	return []byte(fmt.Sprintf("%s-%s-%s", channel, user, timestamp))
}

func (b *markovBot) getMessagesForChannel(channel string) ([]string, error) {
	return b.getMessagesWithPrefix(fmt.Sprintf("%s-", channel))
}

func (b *markovBot) getMessagesForChannelUser(channel, user string) ([]string, error) {
	return b.getMessagesWithPrefix(fmt.Sprintf("%s-%s", channel, user))
}

func (b *markovBot) getMessagesWithPrefix(prefix string) ([]string, error) {
	var messages []string
	iter := b.db.NewIterator(leveldb_util.BytesPrefix([]byte(prefix)), nil)
	defer iter.Release()

	for iter.Next() {
		messages = append(messages, string(iter.Value()))
	}
	return messages, iter.Error()
}

func (b *markovBot) lastTimestampForChannel(channelId string) (float64, error) {
	iter := b.db.NewIterator(leveldb_util.BytesPrefix([]byte(channelId)), nil)
	defer iter.Release()

	var maxTs float64 = 0
	for iter.Next() {
		tokens := strings.Split(string(iter.Key()), "-")
		ts, err := strconv.ParseFloat(tokens[2], 64)
		if err != nil {
			return 0, err
		}
		maxTs = math.Max(maxTs, ts)
	}
	return maxTs, nil
}

func (b *markovBot) addMessage(text, channel, user, timestamp string) {
	key := makeDbKey(channel, user, timestamp)
	if err := b.db.Put(key, []byte(text), nil); err != nil {
		log.Printf("Found error when indexing %v -> %v: %v\n", key, text, err)
	}
}
