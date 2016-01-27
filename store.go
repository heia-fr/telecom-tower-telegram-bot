package main

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/heia-fr/telecom-tower/rollrenderer"
	"log"
)

var store struct {
	db *bolt.DB
}

func openDB(name string) error {
	log.Println(name)
	db, err := bolt.Open(name, 0600, nil)
	if err != nil {
		return err
	}
	store.db = db

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("Message"))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	return err
}

func closeDB() {
	store.db.Close()
}

func saveMessage(m rollrenderer.TextMessage) {
	store.db.Update(func(tx *bolt.Tx) error {
		buf, err := json.Marshal(m)
		if err != nil {
			return err
		}
		b := tx.Bucket([]byte("Message"))
		err = b.Put([]byte("last"), buf)
		return err
	})
}

func loadMessage() (res rollrenderer.TextMessage, err error) {
	err = store.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("Message"))
		v := b.Get([]byte("last"))
		return json.Unmarshal(v, &res)
	})
	return
}
