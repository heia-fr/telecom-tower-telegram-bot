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
// 2016-07-26 | JS | Add Firebase Database

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
	"encoding/json"
	"flag"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/heia-fr/telecom-tower/rollrenderer"
	"github.com/nats-io/nats"
	"github.com/tucnak/telebot"
	"github.com/vharitonsky/iniflags"
	"github.com/BlueMasters/firebasedb"
	"net/http"
	"strings"
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
	pingPeriod   = 30 * time.Second
	maxMsgLen    = 64 // The tower can show up to 21 characters (using 6x8 font).
	maxSenderLen = 32 // The tower can show up to 21 characters (using 6x8 font).
)

// the bot will send notifications to these channels
var notificationChannels = [...]string{"telecom_tower_notifications"}

var bot *telebot.Bot
var sessions = make(map[string]*session) // the key is the telegram chat ID and the user ID
var fbase firebasedb.Reference

func dispatchMessage(sender, text, color string) {
	msg := rollrenderer.TextMessage{
		Introduction: []rollrenderer.Line{
			{Text: "", Font: 6, Color: "#000000"}},
		Conclusion: []rollrenderer.Line{
			{Text: " // ", Font: 6, Color: "#0000FF"}},
		Separator: []rollrenderer.Line{{
			Text: "  --  ", Font: 6, Color: "#FFFFFF"}},
	}

	if sender == "" {
		msg.Body = []rollrenderer.Line{
			{Text: text, Font: 6, Color: color},
		}
	} else {
		msg.Body = []rollrenderer.Line{
			{
				Text: fmt.Sprintf("%s says: ", sender),
				Font: 6, Color: "#FFFFFF"},
			{
				Text: text, Font: 6, Color: color},
		}
	}

	saveMessage(msg)
	bitmap := rollrenderer.RenderMessage(&msg)

	if err := fbase.Ref("currentMessage").Set(&msg); err != nil {
		log.Info(err)
	}

	if err := fbase.Ref("currentBitmap").Set(&bitmap); err != nil {
		log.Info(err)
	}

	natsClient.conn.Publish(natsClient.subject, &bitmap)
}

func publish(w http.ResponseWriter, r *http.Request) {
	var msg struct {
		Sender  string `json:"sender"`
		Message string `json:"message"`
		Color   string `json:"color"`
	}
	err := json.NewDecoder(r.Body).Decode(&msg)
	if err != nil {
		http.Error(w, err.Error(), 403)
		return
	}
	if len(msg.Sender) > maxSenderLen {
		http.Error(w, "Sender too long", 403)
		return
	}
	if len(msg.Message) > maxMsgLen {
		http.Error(w, "Message too long", 403)
		return
	}
	color, ok := colorNames[strings.ToLower(msg.Color)]
	if !ok {
		http.Error(w, "Invalid color", 403)
		return
	}
	dispatchMessage(msg.Sender, msg.Message, color)
}

func stream(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true },
	}

	log.Debugln("Start streaming")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Errorf("Error upgrading connection: %s", err)
		return
	}

	defer conn.Close()

	// readloop, just ignore all incoming websocket messages
	go func(c *websocket.Conn) {
		for {
			if _, _, err := c.NextReader(); err != nil {
				log.Errorf("Read error: %s. Closing", err)
				c.Close()
				break
			}
		}
	}(conn)

	if strings.ToLower(r.FormValue("skip")) != "true" {
		msg, err := loadMessage()
		if err == nil {
			log.Debugln("Sending initial message")
			conn.WriteJSON(rollrenderer.RenderMessage(&msg))
		}
	}

	var messageChannel = make(chan *rollrenderer.BitmapMessage)
	natsClient.conn.BindRecvChan(natsClient.subject, messageChannel)

	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()

mainLoop:
	for {
		select {
		case message := <-messageChannel:
			log.Debugln("Sending message")
			if err := conn.WriteJSON(message); err != nil {
				log.Errorf("Error encoding message: %s", err)
				break mainLoop
			}
		case <-ticker.C:
			log.Debugln("Ping")
			if err := conn.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				log.Errorf("Error sending ping: %s", err)
				break mainLoop
			}
		}
	}

	log.Debugln("End of streaming")
}

var silent *bool

func main() {
	var debug = flag.Bool("debug", false, "Debug mode")
	silent = flag.Bool("silent", false, "Silent")
	var token = flag.String("telegram-token", "", "Telegram Token")
	var natsURL = flag.String("nats-url", nats.DefaultURL, "NATS URL")
	var natsSubject = flag.String("nats-subject", "telecom-tower", "NATS Subject")
	var dbName = flag.String("database", "./database.bolt", "Bolt database name")
	var httpPort = flag.String("port", "8100", "Server port")
	var fireBaseURL = flag.String("firebase-url", "https://telecom-tower.firebaseio.com", "Firebase URL")
	var fireBaseToken = flag.String("firebase-token", "", "Firebase Token")

	iniflags.Parse()

	if *debug {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.InfoLevel)
	}

	log.Infoln("Starting Telegram bot")
	if err := openDB(*dbName); err != nil {
		log.Fatalf("Error opening Bolt database: %s", err)
	}

	var err error
	if err := openNats(*natsURL, *natsSubject); err != nil {
		log.Fatalf("Error connecting to NATS server: %s", err)

	}

	// Connect to Firebase Database
	fbase = firebasedb.NewReference(*fireBaseURL)
	if fbase.Error != nil {
		log.Fatalf("Error opening firebase: %s", err)
	}
	fbase = fbase.Auth(firebasedb.Secret{Token:*fireBaseToken})

	bot, err = telebot.NewBot(*token)
	if err != nil {
		log.Fatalf("Error creating bot: %s", err)
	}

	go processTelegramMessages(bot)

	r := mux.NewRouter()
	r.HandleFunc("/stream", stream)
	r.HandleFunc("/publish", publish).Methods("POST")
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./html/")))

	http.Handle("/", r)
	http.ListenAndServe(fmt.Sprintf(":%s", *httpPort), nil)

	log.Infoln("Terminated.")
}
