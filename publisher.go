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
	"fmt"
	"github.com/nats-io/nats"
	"log"
)

type Line struct {
	Text  string `json:"text"`
	Font  int    `json:"font"`
	Color string `json:"color"`
}

type RollingMessage struct {
	Body         []Line `json:"body"`
	Introduction []Line `json:"introduction"`
	Conclusion   []Line `json:"conclusion"`
	Separator    []Line `json:"separator"`
}

var natsConn *nats.EncodedConn
var natsSubject string

func initPublisher(url string, subject string) {
	nc, err := nats.Connect(url)
	if err != nil {
		log.Fatal(err)
	}

	natsConn, err = nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(err)
	}
	natsSubject = subject
}

func (s *session) publishMessage(text string) {
	towerMessage := RollingMessage{
		Body: []Line{
			Line{
				Text:  fmt.Sprintf("%s says: ", s.sender.FirstName),
				Font:  6,
				Color: "#FFFFFF"},
			Line{
				Text:  text,
				Font:  6,
				Color: s.color},
		},
		Introduction: []Line{Line{Text: "", Font: 6, Color: "#000000"}},
		Conclusion:   []Line{Line{Text: " // ", Font: 6, Color: "#0000FF"}},
		Separator:    []Line{Line{Text: "  --  ", Font: 6, Color: "#FFFFFF"}},
	}

	if s.anonymous {
		towerMessage.Body = []Line{
			Line{Text: text, Font: 6, Color: s.color},
		}
	} else {
		towerMessage.Body = []Line{
			Line{
				Text: fmt.Sprintf("%s says: ", s.sender.FirstName),
				Font: 6, Color: "#FFFFFF"},
			Line{
				Text: text, Font: 6, Color: s.color},
		}
	}

	natsConn.Publish(natsSubject, towerMessage)
}
