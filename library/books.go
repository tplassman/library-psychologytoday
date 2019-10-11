package library

import (
	"encoding/binary"
	"encoding/json"

	"github.com/boltdb/bolt"
)

const BOOKS_BUCKET = "books"

type BookRepo interface {
	All() ([]*book, error)
	One(id int) (*book, error)
	New(title, author, isnb, description string) (*book, error)
	Update(b *book) error
	Delete(id int) error
	CheckIn(id int) error
	CheckOut(id int) error
}

func NewBookRepo(db *bolt.DB) (BookRepo, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BOOKS_BUCKET))

		return err
	})
	if err != nil {
		return nil, err
	}

	r := bookStore{db}

	return r, nil
}

type book struct {
	ID          int
	Title       string
	Author      string
	ISBN        string
	Description string
	CheckedOut  bool
}

func (b *book) IsCheckedOut() bool {
	// TODO: Replace with check against event stream
	return b.CheckedOut
}

type bookStore struct {
	db *bolt.DB
}

func (r bookStore) All() ([]*book, error) {
	var bs []*book

	err := r.db.View(func(tx *bolt.Tx) error {
		bb := tx.Bucket([]byte(BOOKS_BUCKET))
		c := bb.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			b := &book{}
			if err := json.Unmarshal(v, b); err != nil {
				return err
			}
			bs = append(bs, b)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return bs, nil
}

func (r bookStore) One(id int) (*book, error) {
	b := &book{}

	err := r.db.View(func(tx *bolt.Tx) error {
		bb := tx.Bucket([]byte(BOOKS_BUCKET))
		v := bb.Get(itob(id))

		return json.Unmarshal(v, b)
	})
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (r bookStore) New(title, author, isbn, description string) (*book, error) {
	var b book

	err := r.db.Update(func(tx *bolt.Tx) error {
		bb := tx.Bucket([]byte(BOOKS_BUCKET))
		id, _ := bb.NextSequence()
		b = book{int(id), title, author, isbn, description, false}

		buf, err := json.Marshal(b)
		if err != nil {
			return err
		}

		return bb.Put(itob(int(id)), buf)
	})
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func (r bookStore) Update(b *book) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bb := tx.Bucket([]byte(BOOKS_BUCKET))

		buf, err := json.Marshal(b)
		if err != nil {
			return err
		}

		return bb.Put(itob(b.ID), buf)
	})
}

func (r bookStore) Delete(id int) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(BOOKS_BUCKET)).Cursor()

		for k, _ := c.Seek(itob(id)); k != nil; k, _ = c.Next() {
			return c.Delete()
		}

		return nil
	})
}

func (r bookStore) CheckIn(id int) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bb := tx.Bucket([]byte(BOOKS_BUCKET))

		// Fetch book by ID and unmarshal
		v := bb.Get(itob(id))
		b := &book{}
		if err := json.Unmarshal(v, b); err != nil {
			return err
		}

		// Set to checked in and marshal back into DB
		b.CheckedOut = false

		// Put marshalled data back to same key
		buf, err := json.Marshal(b)
		if err != nil {
			return err
		}

		return bb.Put(itob(b.ID), buf)
	})
}

func (r bookStore) CheckOut(id int) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bb := tx.Bucket([]byte(BOOKS_BUCKET))

		// Fetch book by ID and unmarshal
		v := bb.Get(itob(id))
		b := &book{}
		if err := json.Unmarshal(v, b); err != nil {
			return err
		}

		// Set to checked out and marshal back into DB
		b.CheckedOut = true

		// Put marshalled data back to same key
		buf, err := json.Marshal(b)
		if err != nil {
			return err
		}

		return bb.Put(itob(b.ID), buf)
	})
}

/**
 * Helper fuction to return an 8-byte big endian representation of v. for querying DB keys
 */
func itob(v int) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}
