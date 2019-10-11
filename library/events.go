package library

import (
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
)

const EVENTS_BUCKET = "events"

type EventRepo interface {
	BookAdded(id int) error
	BookRemoved(id int) error
	BookCheckedIn(id int) error
	BookCheckedOut(id int) error
}

func NewEventRepo(db *bolt.DB) (EventRepo, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(EVENTS_BUCKET))

		return err
	})
	if err != nil {
		return nil, err
	}

	r := eventStore{db}

	return r, nil
}

type event struct {
	Type   string
	BookID int
}

func writeEvent(db *bolt.DB, e *event) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_BUCKET))

		// Marshal event data into bytes.
		buf, err := json.Marshal(e)
		if err != nil {
			return err
		}

		return b.Put([]byte(time.Now().Format(time.RFC3339)), buf)
	})
}

type eventStore struct {
	db *bolt.DB
}

func (r eventStore) BookAdded(id int) error {
	return writeEvent(r.db, &event{"BOOK_ADDED", id})
}
func (r eventStore) BookRemoved(id int) error {
	return writeEvent(r.db, &event{"BOOK_REMOVED", id})
}
func (r eventStore) BookCheckedIn(id int) error {
	return writeEvent(r.db, &event{"BOOK_CHECKED_IN", id})
}
func (r eventStore) BookCheckedOut(id int) error {
	return writeEvent(r.db, &event{"BOOK_CHECKED_OUT", id})
}
