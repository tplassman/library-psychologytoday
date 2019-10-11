package library

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/boltdb/bolt"
)

const EVENTS_BUCKET = "events"

const (
	EVENT_BOOK_ADDED       = "BOOK_ADDED"
	EVENT_BOOK_REMOVED     = "BOOK_REMOVED"
	EVENT_BOOK_CHECKED_IN  = "BOOK_CHECKED_IN"
	EVENT_BOOK_CHECKED_OUT = "BOOK_CHECKED_OUT"
)

type EventRepo interface {
	All() ([]*event, error)
	AllForBook(id int) ([]*event, error)
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
	Time   time.Time
	Type   string
	BookID int
}

func (e *event) Title() string {
	switch t := e.Type; t {
	case EVENT_BOOK_ADDED:
		return "Book added to library"
	case EVENT_BOOK_REMOVED:
		return "Book removed from library"
	case EVENT_BOOK_CHECKED_IN:
		return "Book checked in to library"
	case EVENT_BOOK_CHECKED_OUT:
		return "Book checked out of library"
	default:
		return "Unknown"
	}
}

func (e *event) PrettyTime() string {
	return e.Time.Format(time.UnixDate)
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

			// Add key timestamp to event
			t, err := time.Parse(time.RFC3339, bytes.NewBuffer(k).String())
			if err != nil {
				return err
			}

			e.Time = t
			es = append(es, e)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return es, nil
}

func (r eventStore) AllForBook(id int) ([]*event, error) {
	var es []*event

	err := r.db.View(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(EVENTS_BUCKET)).Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			e := &event{}
			if err := json.Unmarshal(v, e); err != nil {
				return err
			}

			if e.BookID != id {
				continue
			}

			// Add key timestamp to event
			t, err := time.Parse(time.RFC3339, bytes.NewBuffer(k).String())
			if err != nil {
				return err
			}
			e.Time = t

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
	return writeEvent(r.db, id, EVENT_BOOK_ADDED)
}
func (r eventStore) BookRemoved(id int) error {
	return writeEvent(r.db, id, EVENT_BOOK_REMOVED)
}
func (r eventStore) BookCheckedIn(id int) error {
	return writeEvent(r.db, id, EVENT_BOOK_CHECKED_IN)
}
func (r eventStore) BookCheckedOut(id int) error {
	return writeEvent(r.db, id, EVENT_BOOK_CHECKED_OUT)
}

func writeEvent(db *bolt.DB, bookId int, eventType string) error {
	return db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(EVENTS_BUCKET))
		t := time.Now()
		e := &event{t, eventType, bookId}

		// Marshal event data into bytes.
		buf, err := json.Marshal(e)
		if err != nil {
			return err
		}

		return b.Put([]byte(t.Format(time.RFC3339)), buf)
	})
}
