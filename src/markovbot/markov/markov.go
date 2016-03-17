package markov

import (
	"math/rand"
	"strings"
)

const (
	stopSentinel    = "<STOP>"
	prefixDelimiter = " "
)

type MarkovModel interface {
	AddMessage(message string)
	GenerateSentence(maxWords int) string
}

type prefix []string

type markovModel struct {
	// map of prefix -> next token; could be more efficient keeping counts
	model map[string][]string
	order int
}

func NewMarkovModel(order int) MarkovModel {
	return &markovModel{
		model: make(map[string][]string),
		order: order,
	}
}

func (p prefix) String() string {
	return strings.Join(p, prefixDelimiter)
}

func (p prefix) Shift(word string) {
	copy(p, p[1:])
	p[len(p)-1] = word
}

func (m *markovModel) addToken(prefix prefix, token string) {
	key := prefix.String()
	m.model[key] = append(m.model[key], token)
}

func (m *markovModel) getNextToken(prefix prefix) string {
	key := prefix.String()
	choices := m.model[key]
	if len(choices) == 0 {
		return stopSentinel
	}
	return choices[rand.Intn(len(choices))]
}

func normalize(token string) string {
	if token != "" {
		// no notifications
		if strings.Contains(token, "@") {
			return ""
		}
		return strings.ToLower(token)
	}
	return ""
}

func (m *markovModel) AddMessage(message string) {
	prefix := make(prefix, m.order)
	numAdded := 0
	for _, token := range strings.Split(message, " ") {
		normalized := normalize(token)
		if normalized != "" {
			// add the current token
			m.addToken(prefix, normalized)

			// shift the prefix
			prefix.Shift(normalized)
			numAdded++
		}
	}

	if numAdded > 0 {
		// add the stop sentinel
		m.addToken(prefix, stopSentinel)
	}
}

func (m *markovModel) GenerateSentence(maxWords int) string {
	prefix := make(prefix, m.order)
	var sentence []string
	for len(sentence) < maxWords {
		next := m.getNextToken(prefix)
		if next == stopSentinel {
			break
		}
		sentence = append(sentence, next)
		prefix.Shift(next)
	}

	if len(sentence) == 0 {
		return "< I've got no seed data for that. :( >"
	}
	return strings.Join(sentence, " ")
}
