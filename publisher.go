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

//
// Telegram bot
//

package main

import (
	"fmt"
	"github.com/nats-io/nats"
)

type Line struct {
	Text  string
	Font  int
	Color string
}

type RollingMessage struct {
	Body         []Line
	Introduction []Line
	Conclusion   []Line
	Separator    []Line
}

type natsGateway struct {
	conn         *nats.EncodedConn
	pubSubject   string
	towerMessage RollingMessage
}

var natsGw natsGateway

func (ng natsGateway) init(url string, pubSubject string, qSubject string) {
	nc, _ := nats.Connect(url)
	ng.conn, _ = nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	ng.pubSubject = pubSubject

	ng.conn.Subscribe(qSubject, func(subj, reply string, msg string) {
		ng.conn.Publish(reply, ng.towerMessage)
	})
}

func (ng natsGateway) publishMessage(s *session) {
	ng.towerMessage = RollingMessage{
		Introduction: []Line{Line{Text: "", Font: 6, Color: "#000000"}},
		Conclusion:   []Line{Line{Text: " // ", Font: 6, Color: "#0000FF"}},
		Separator:    []Line{Line{Text: "  --  ", Font: 6, Color: "#FFFFFF"}},
	}

	if s.anonymous {
		ng.towerMessage.Body = []Line{
			Line{Text: s.message.Text, Font: 6, Color: s.color},
		}
	} else {
		ng.towerMessage.Body = []Line{
			Line{
				Text: fmt.Sprintf("%s says: ", s.sender.FirstName),
				Font: 6, Color: "#FFFFFF"},
			Line{
				Text: s.message.Text, Font: 6, Color: s.color},
		}
	}

	ng.conn.Publish(ng.pubSubject, ng.towerMessage)
}

func init() {
	natsGw.towerMessage = RollingMessage{
		Body: []Line{Line{
			Text: "Telecom Tower © 2016 HEIA-FR",
			Font: 6, Color: "#FFA500"}},
		Introduction: []Line{Line{Text: "", Font: 6, Color: "#000000"}},
		Conclusion:   []Line{Line{Text: " // ", Font: 6, Color: "#0000FF"}},
		Separator:    []Line{Line{Text: "  --  ", Font: 6, Color: "#FFFFFF"}},
	}
}
