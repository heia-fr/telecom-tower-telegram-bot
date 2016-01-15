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
// 2016-01-12 | JS | Last version (add status command)

//
// Telegram bot
//

package main

import (
	"fmt"
	"github.com/tucnak/telebot"
)

const (
	maxMsgLen = 64 // The tower can show up to 21 characters (using 6x8 font).
)

var colorKeyboard = telebot.SendOptions{
	ReplyMarkup: telebot.ReplyMarkup{
		CustomKeyboard: [][]string{
			[]string{"Red", "SkyBlue"},
			[]string{"Green", "Orange"},
			[]string{"Gold", "HotPink"},
		},
		OneTimeKeyboard: true,
	},
}

var msgKeyboard = telebot.SendOptions{
	ReplyMarkup: telebot.ReplyMarkup{
		CustomKeyboard: [][]string{
			[]string{"I ♥︎ Computer Science"},
			[]string{"I ♥︎ Telecommunications"},
			[]string{"I ♥︎ HEIA-FR"},
		},
		OneTimeKeyboard: true,
	},
}

var hideKeyboard = telebot.SendOptions{
	ReplyMarkup: telebot.ReplyMarkup{
		HideCustomKeyboard: true,
	},
}

func (s *session) sayHello() {
	bot.SendMessage(s.conversation,
		fmt.Sprintf(
			"Hello %s, nice to see you. Please, enter the color for your message.",
			s.sender.FirstName),
		&colorKeyboard)
}

func (s *session) sayStatus() {
	bot.SendMessage(s.conversation, lastMessage, &hideKeyboard)
}

func (s *session) sayCanceled() {
	bot.SendMessage(s.conversation, "OK", &hideKeyboard)
}

func (s *session) sayBadColor() {
	bot.SendMessage(s.conversation,
		"I don't know this color. Please try another one.",
		&colorKeyboard)
}

func (s *session) sayGoodColor() {
	bot.SendMessage(s.conversation,
		fmt.Sprintf("Good. Now please enter your message (max %d characters).", maxMsgLen),
		&msgKeyboard)
}

func (s *session) sayTooLongText() {
	bot.SendMessage(s.conversation,
		"Your message is too long. Please try a shorter one.",
		&msgKeyboard)
}

func (s *session) sayGoodText() {
	bot.SendMessage(s.conversation,
		"Thank you. I will display your message soon.",
		&hideKeyboard)
}
