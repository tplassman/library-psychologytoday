package library

import (
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
)

const EVENTS_BUCKET = "events"

type EventRepo interface {
	All() ([]*event, error)
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

type eventStore struct {
	db *bolt.DB
}

func (r eventStore) All() ([]*event, error) {
	var es []*event

	err := r.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(EVENTS_BUCKET)).Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			e := &event{}
			if err := json.Unmarshal(v, e); err != nil {
				return err
			}
			es = append(es, e)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return es, nil
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
