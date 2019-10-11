package main

import (
	"log"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/tplassman/ptstacks/library"
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
	bookRepo, err := library.NewBookRepo(db)
	if err != nil {
		panic(err)
	}
	eventRepo, err := library.NewEventRepo(db)
	if err != nil {
		panic(err)
	}

	// Create server and attach routes
	s := library.Server{bookRepo, eventRepo, mux.NewRouter(), env}
	s.Routes()

	// Start server
	secret := []byte(env["SECRET_KEY"])
	secure := csrf.Secure(env["ENVIRONMENT"] != "development")
	log.Fatal(http.ListenAndServe(":8080", csrf.Protect(secret, secure)(s.Router)))
}
