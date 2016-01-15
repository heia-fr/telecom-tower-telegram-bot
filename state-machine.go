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
// 2016-01-12 | JS | Last change

//
// Telegram bot
//

package main

import (
	"fmt"
	"github.com/tucnak/telebot"
	"log"
	"strings"
)

const (
	printCommand          = "/print"
	anonymousPrintCommand = "/xprint"
	cancelCommand         = "/cancel"
	statusCommand         = "/status"
)

var lastMessage string = "(no message)"

type stateFn func(*session)

func idleState(s *session) {
	if s.message.Text == printCommand || s.message.Text == anonymousPrintCommand {
		s.sayHello()
		s.state = checkColorState
		s.anonymous = s.message.Text == anonymousPrintCommand
	} else if s.message.Text == statusCommand {
		s.sayStatus()
	}
}

func checkColorState(s *session) {
	if s.message.Text == cancelCommand {
		s.sayCanceled()
		s.state = idleState
		return
	}
	color, ok := colorNames[strings.ToLower(s.message.Text)]
	if ok {
		s.sayGoodColor()
		s.color = color
		s.state = checkTextState
	} else {
		s.sayBadColor()
	}
}

func checkTextState(s *session) {
	if s.message.Text == cancelCommand {
		s.sayCanceled()
		s.state = idleState
		return
	}
	if len(s.message.Text) <= maxMsgLen {
		log.Printf(
			"%s %s (%s/%d) says : \"%s\" in %s",
			s.sender.FirstName, s.sender.LastName, s.sender.Username, s.sender.ID,
			s.message.Text, s.color)
		lastMessage = fmt.Sprintf("%s says : \"%s\"", s.sender.FirstName, s.message.Text)
		// Send a notification to channels
		for _, username := range notificationChannels {
			err := bot.SendMessage(
				telebot.Chat{Type: "channel", Username: username},
				fmt.Sprintf("%s says : \"%s\"", s.sender.FirstName, s.message.Text),
				nil)
			if err != nil {
				log.Printf("Error sending notification: %s\n", err)
			}
		}
		s.sayGoodText()
		natsGw.publishMessage(s)
		s.state = idleState
	} else {
		s.sayTooLongText()
	}
}
