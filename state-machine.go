// Copyright 2015 Jacques Supcik, HEIA-FR
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

//
// Telegram bot
//

package main

import (
	"log"
	"strings"
)

const (
	printCommand          = "/print"
	anonymousPrintCommand = "/xprint"
	cancelCommand         = "/cancel"
)

type stateFn func(*session)

func idleState(s *session) {
	if s.lastMessage.Text == printCommand || s.lastMessage.Text == anonymousPrintCommand {
		s.sayHello()
		s.state = checkColorState
		s.anonymous = s.lastMessage.Text == anonymousPrintCommand
	}
}

func checkColorState(s *session) {
	if s.lastMessage.Text == cancelCommand {
		s.sayCanceled()
		s.state = idleState
		return
	}
	color, ok := colorNames[strings.ToLower(s.lastMessage.Text)]
	if ok {
		s.sayGoodColor()
		s.color = color
		s.state = checkTextState
	} else {
		s.sayBadColor()
	}
}

func checkTextState(s *session) {
	if s.lastMessage.Text == cancelCommand {
		s.sayCanceled()
		s.state = idleState
		return
	}
	if len(s.lastMessage.Text) <= maxMsgLen {
		log.Printf(
			"%s %s (%s) says : \"%s\" in %s",
			s.sender.FirstName, s.sender.LastName, s.sender.Username,
			s.lastMessage.Text, s.color)
		s.sayGoodText()
		s.publishMessage(s.lastMessage.Text)
		s.state = idleState
	} else {
		s.sayTooLongText()
	}
}
