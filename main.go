package main

import (
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/tplassman/library-psychologytoday/library"
)

const DB_NAME = "data.db"

func main() {
	// Load env file
	env, err := godotenv.Read()
	if err != nil {
		panic(err)
	}

	// Open DB and create required repositories
	db, err := bolt.Open(DB_NAME, 0600, nil)
	if err != nil {
		panic(err)
	}
	booksRepo, err := library.NewBooksRepo(db)
	if err != nil {
		panic(err)
	}
	eventsRepo, err := library.NewEventsRepo(db)
	if err != nil {
		panic(err)
	}

	// Create server and attach routes
	s := library.Server{booksRepo, eventsRepo, mux.NewRouter(), env}
	s.Routes()

	// Start server
	secret := []byte(env["SECRET_KEY"])
	secure := csrf.Secure(env["ENVIRONMENT"] != "development")
	log.Fatal(http.ListenAndServe(":8080", csrf.Protect(secret, secure)(s.Router)))
}

/**
 * Helper fuction to return an 8-byte big endian representation of v. for querying DB keys
 */
func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)

	return b
}
