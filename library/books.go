package library

import (
	"encoding/json"

	"github.com/boltdb/bolt"
)

const BOOKS_BUCKET = "books"

type BooksRepo interface {
	All() ([]*book, error)
	One(id int) (*book, error)
	New(title, author, isnb, description string) (*book, error)
	Delete(id int) error
	CheckIn(id int) error
	CheckOut(id int) error
}

func NewBooksRepo(db *bolt.DB) (BooksRepo, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(BOOKS_BUCKET))

		return err
	})
	if err != nil {
		return nil, err
	}

	r := booksRepo{db}

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

type booksRepo struct {
	db *bolt.DB
}

func (r booksRepo) All() ([]*book, error) {
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

func (r booksRepo) One(id int) (*book, error) {
	b := &book{}

	err := r.db.View(func(tx *bolt.Tx) error {
		bb := tx.Bucket([]byte(BOOKS_BUCKET))
		v := bb.Get(itob(uint64(id)))

		return json.Unmarshal(v, b)
	})
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (r booksRepo) New(title, author, isbn, description string) (*book, error) {
	var b book

	err := r.db.Update(func(tx *bolt.Tx) error {
		bb := tx.Bucket([]byte(BOOKS_BUCKET))
		id, _ := bb.NextSequence()
		b = book{int(id), title, author, isbn, description, false}

		buf, err := json.Marshal(b)
		if err != nil {
			return err
		}

		return bb.Put(itob(id), buf)
	})
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func (r booksRepo) Delete(id int) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		c := tx.Bucket([]byte(BOOKS_BUCKET)).Cursor()

		for k, _ := c.Seek(itob(uint64(id))); k != nil; k, _ = c.Next() {
			return c.Delete()
		}

		return nil
	})
}

func (r booksRepo) CheckIn(id int) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bb := tx.Bucket([]byte(BOOKS_BUCKET))
		c := bb.Cursor()
		b := &book{}

		// Fetch book by ID and unmarshal
		for k, v := c.Seek(itob(uint64(id))); k != nil; k, v = c.Next() {
			if err := json.Unmarshal(v, b); err != nil {
				return err
			}
		}

		// Set to checked in and marshal back into DB
		b.CheckedOut = false

		buf, err := json.Marshal(b)
		if err != nil {
			return err
		}

		return bb.Put(itob(uint64(id)), buf)
	})
}

func (r booksRepo) CheckOut(id int) error {
	return r.db.Update(func(tx *bolt.Tx) error {
		bb := tx.Bucket([]byte(BOOKS_BUCKET))
		c := bb.Cursor()
		b := &book{}

		// Fetch book by ID and unmarshal
		for k, v := c.Seek(itob(uint64(id))); k != nil; k, v = c.Next() {
			if err := json.Unmarshal(v, b); err != nil {
				return err
			}
		}

		// Set to checked out and marshal back into DB
		b.CheckedOut = true

		buf, err := json.Marshal(b)
		if err != nil {
			return err
		}

		return bb.Put(itob(uint64(id)), buf)
	})
}
