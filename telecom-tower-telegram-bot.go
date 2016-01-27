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

// Infos for BotFather
/*

/setcommands
print  - print message on tower
cancel - cancel current operation
status - display current message

/setdescription
This bot displays your message on the LED display of the telecom tower of the "Haute école d'ingénierie et d'architecture Fribourg"
Source code available on github: https://github.com/heia-fr/telecom-tower-telegram-bot

/setabouttext
By Jacques Supcik - Haute école d'ingénierie et d'architecture Fribourg - Filière télécommunications
*/

//
// Telegram bot
//

package main

import (
	"flag"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/heia-fr/telecom-tower/rollrenderer"
	"github.com/nats-io/nats"
	"github.com/tucnak/telebot"
	"github.com/vharitonsky/iniflags"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type session struct {
	state        stateFn
	color        string
	anonymous    bool
	conversation telebot.Chat
	sender       telebot.User
	message      telebot.Message
}

const (
	longPollTime = 300 * time.Second
)

var notificationChannels = [...]string{"telecom_tower_notifications"}

var bot *telebot.Bot
var sessions = make(map[string]*session) // the key is the telegram chat ID and the user ID

func home(w http.ResponseWriter, r *http.Request) {
	f, _ := os.Open("./html/index.html")
	io.Copy(w, f)
}

func stream(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// readloop, just ignore all incoming websocket messages
	go func(c *websocket.Conn) {
		for {
			if _, _, err := c.NextReader(); err != nil {
				c.Close()
				break
			}
		}
	}(conn)

	msg, err := loadMessage()
	if err == nil {
		log.Println(rollrenderer.RenderMessage(&msg))
		conn.WriteJSON(rollrenderer.RenderMessage(&msg))
	}

	var messageChannel = make(chan *rollrenderer.BitmapMessage)
	natsClient.conn.BindRecvChan(natsClient.subject, messageChannel)

	for {
		message := <-messageChannel
		log.Println(message)
		if err := conn.WriteJSON(message); err != nil {
			return
		}
	}
}

func main() {
	var token = flag.String("telegram-token", "", "Telegram Token")
	var natsURL = flag.String("nats-url", nats.DefaultURL, "NATS URL")
	var natsSubject = flag.String("nats-subject", "telecom-tower", "NATS Subject")
	var dbName = flag.String("database", "./database.bolt", "Bolt database name")

	iniflags.Parse()

	if err := openDB(*dbName); err != nil {
		log.Fatal(err)
	}

	var err error
	openNats(*natsURL, *natsSubject)

	bot, err = telebot.NewBot(*token)
	if err != nil {
		log.Fatal(err)
	}

	go processTelegramMessages(bot)

	r := mux.NewRouter()
	r.HandleFunc("/", home)
	r.HandleFunc("/stream", stream)

	r.PathPrefix("/static/").Handler(
		http.StripPrefix(
			"/static/", http.FileServer(http.Dir("./static/"))))

	http.Handle("/", r)
	http.ListenAndServe(":8100", nil)

}
