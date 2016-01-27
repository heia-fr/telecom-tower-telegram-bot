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

package main

import (
	_ "fmt"
	"github.com/nats-io/nats"
)

var natsClient struct {
	conn    *nats.EncodedConn
	subject string
}

func openNats(url string, subject string) {
	nc, _ := nats.Connect(url)
	natsClient.conn, _ = nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	natsClient.subject = subject
}

func closeNats() {
	natsClient.conn.Close()
}

// 	ng.conn.Publish(ng.pubSubject, ng.towerMessage)
