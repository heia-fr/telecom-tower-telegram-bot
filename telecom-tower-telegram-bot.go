// Copyright 2016 Jacques Supcik, HEIA-FR
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// 2015-12-29 | JS | First version
// 2016-01-12 | JS | Last edit

//
// Telegram bot
//

package main

import (
	"flag"
	"fmt"
	"github.com/nats-io/nats"
	"github.com/tucnak/telebot"
	"github.com/vharitonsky/iniflags"
	"log"
	"time"
)

type session struct {
	state        stateFn
	color        string
	anonymous    bool
	conversation telebot.Chat
	sender       telebot.User
	lastMessage  telebot.Message
}

const (
	longPollTime = 300 * time.Second
)

var notificationChannels = [...]string{"telecom_tower_notifications"}

var bot *telebot.Bot
var sessions = make(map[string]*session) // the key is the telegram chat ID and the user ID

func main() {
	var token = flag.String("telegram-token", "", "Telegram Token")
	var natsUrl = flag.String("nats-url", nats.DefaultURL, "NATS URL")
	var natsSubject = flag.String("nats-subject", "heiafr.telecomtower.bot", "NATS Subject")
	iniflags.Parse()

	var err error
	bot, err = telebot.NewBot(*token)
	if err != nil {
		log.Fatal(err)
	}

	messages := make(chan telebot.Message)
	bot.Listen(messages, longPollTime)

	initPublisher(*natsUrl, *natsSubject)

	for message := range messages {
		key := fmt.Sprintf("%x:%x", message.Chat.ID, message.Sender.ID)
		currentSession, ok := sessions[key]
		if !ok {
			currentSession = new(session)
			currentSession.state = idleState
			currentSession.conversation = message.Chat
			currentSession.sender = message.Sender
			sessions[key] = currentSession
		}
		currentSession.lastMessage = message
		currentSession.state(currentSession)
	}

}
