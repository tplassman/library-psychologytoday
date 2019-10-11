package library

import (
	"encoding/json"

	"github.com/boltdb/bolt"
)

const EVENTS_BUCKET = "events"

type EventsRepo interface {
	BookAdded(id int) error
	BookRemoved(id int) error
	BookCheckedIn(id int) error
	BookCheckedOut(id int) error
}

func NewEventsRepo(db *bolt.DB) (EventsRepo, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(EVENTS_BUCKET))

		return err
	})
	if err != nil {
		return nil, err
	}

	r := eventsRepo{db}

	return r, nil
}

type event struct {
	Type   string
	BookID int
}

func writeEvent(db *bolt.DB, e *event) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_BUCKET))
		id, _ := b.NextSequence()

		// Marshal event data into bytes.
		buf, err := json.Marshal(e)
		if err != nil {
			return err
		}

		return b.Put(itob(id), buf)
	})
}

type eventsRepo struct {
	db *bolt.DB
}

func (r eventsRepo) BookAdded(id int) error {
	return writeEvent(r.db, &event{"BOOK_ADDED", id})
}
func (r eventsRepo) BookRemoved(id int) error {
	return writeEvent(r.db, &event{"BOOK_REMOVED", id})
}
func (r eventsRepo) BookCheckedIn(id int) error {
	return writeEvent(r.db, &event{"BOOK_CHECKED_IN", id})
}
func (r eventsRepo) BookCheckedOut(id int) error {
	return writeEvent(r.db, &event{"BOOK_CHECKED_OUT", id})
}
