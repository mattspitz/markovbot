package main

import (
	"flag"
	"log"
	"markovbot"
)

func main() {
	var flagApiToken string
	var flagDbFilename string
	var flagMarkovOrder int
	var flagMaxResponseWords int
	var flagMaxAgeHours int

	flag.StringVar(
		&flagApiToken,
		"markovbot.api_token",
		"",
		"Slack API token for fetching results [REQUIRED]",
	)
	flag.StringVar(
		&flagDbFilename,
		"markovbot.db_filename",
		"./db",
		"Filename for LevelDB database",
	)
	flag.IntVar(
		&flagMarkovOrder,
		"markovbot.markov_order",
		1,
		"Order for the Markov model",
	)
	flag.IntVar(
		&flagMaxResponseWords,
		"markovbot.max_response_words",
		20,
		"Maximum number of words for a response",
	)
	flag.IntVar(
		&flagMaxAgeHours,
		"markovbot.max_age_hours",
		24*7*12,
		"Maximum age of chats to retain per-channel (default 12 weeks)",
	)

	flag.Parse()

	if flagApiToken == "" {
		log.Fatal("markovbot.api_token is required!")
	}
	if flagMaxAgeHours <= 0 {
		log.Fatal("markovbot.max_age_hours must be positive!")
	}

	s, err := markovbot.NewMarkovBot(
		flagApiToken, flagDbFilename, flagMarkovOrder,
		flagMaxResponseWords, flagMaxAgeHours,
	)
	if err != nil {
		log.Panic(err)
	}
	s.Start()
}
