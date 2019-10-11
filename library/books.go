package library

import (
	"encoding/json"

	"github.com/boltdb/bolt"
)

var booksBucket = "books"

type BooksRepo interface {
	All() ([]*book, error)
	One(id int) (*book, error)
	New(title, author, isnb, description string) (*book, error)
	Delete(id int) error
}

func NewBooksRepo(db *bolt.DB) (BooksRepo, error) {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(booksBucket))

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
}

func (b *book) IsCheckedOut() bool {
	return false
}

type booksRepo struct {
	db *bolt.DB
}

func (r booksRepo) All() ([]*book, error) {
	var bs []*book

	err := r.db.View(func(tx *bolt.Tx) error {
		bb := tx.Bucket([]byte(booksBucket))
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
		bb := tx.Bucket([]byte(booksBucket))
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
		bb := tx.Bucket([]byte(booksBucket))
		id, _ := bb.NextSequence()
		b = book{int(id), title, author, isbn, description}

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
		c := tx.Bucket([]byte(booksBucket)).Cursor()
		for k, _ := c.Seek(itob(uint64(id))); k != nil; k, _ = c.Next() {
			return c.Delete()
		}

		return nil
	})
}

// func (b *book) checkIn() error {
// 	b.CheckedOut = false
//
// 	if err := b.save(false); err != nil {
// 		return err
// 	}
//
// 	return bookCheckedIn(b.ID)
// }

// func (r booksRepo) checkOut() error {
// 	b.CheckedOut = true
//
// 	if err := b.save(false); err != nil {
// 		return err
// 	}
//
// 	return bookCheckedOut(b.ID)
// }
