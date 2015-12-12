# Markovbot

Do you ever get the feeling that chat rooms are all the same? That you just keep reading the same thing over and over and over again? Remix it with Markovbot!

Markovbot sits in whatever channels you invite it to (private or group) and, when, asked, generates messages based on the content that's been posted to those channels using [Markov chains](https://en.wikipedia.org/wiki/Markov_chain).

Markovbot doesn't require external internet access, but you do need to run a server somewhere.

All methods require a Slack API key. You can get one by creating a Slackbot, which can later post messages to various channels! Get started here: `https://<YOUR-DOMAIN>.slack.com/services/new/bot`

## Usage in Slack

Markovbot can generate messages based on channel content or content from one or more users.

First, once the Markovbot server is up and connected to your Slack instance, invite it to a room using `/invite <bot-username>`. Markovbot can live in multiple rooms.

When the bot is in a room, typing "markovbot" triggers a new message. If there are usernames following "markovbot", it will try to use those usernames to generate content.

`markovbot` generates content based on the whole channel

`markovbot: bill sandy andy` generates content based on the chat history from users bill, sandy, and andy.


## Server

It's Go, so just build the binary and you're off to the races. Here are the command options you can specify:
```
Usage of ./main:
  -markovbot.api_token="": Slack API token for fetching results [REQUIRED]
  -markovbot.db_filename="./db": Filename for LevelDB database
  -markovbot.markov_order=1: Order for the Markov model
  -markovbot.max_age_hours=2016: Maximum age of chats to retain per-channel (default 12 weeks)
  -markovbot.max_response_words=20: Maximum number of words for a response
```

Hilarity ensues!
