# Psychology Today Library Application

CRUD app for books in a library. Funcitonality included:

1. Adding new books to the stacks
	* Title, Author, ISBN, Description
2. Editing existing books
3. Removing books from the stacks
4. Checking books in and out
	* Bookes that are checked our cannot be edited

## Installation

### Requirements

* [Go](https://golang.org/dl/) v1.11 or higher
* [NodeJS](https://nodejs.org/en/) v8.16.0 or higher (development only)
* [npm](https://www.npmjs.com/get-npm) v6.4.1 or higher (development only)

### Server Build Process

1. Copy `.example-env` to `.env`.

    ```
    $ cp .env-example .env
    ```

2. Add 32 byte long key to .env file for CSRF tokens.

3. Build or run `main.go`.

    ```
    $ go build main.go
    $ ./main
    ```

    or

    ```
    $ go run main.go
    ```

4. Visit `localhost:8080` in web browser.

### Assets Builld Process (development only)

Styles and scripts are built from `assets` directory into `static` directory.

1. Install front end dependencies.

    ```
    $ npm install
    ```

2. Build or watch assets.

    ```
    $ gulp build
    ```

    or

    ```
    $ gulp watch
    ```

## Features

* Add new book consisting of title, author, ISBN, and description.
* View all books in the library
* View log of all events on a book
* Edit details of existing book
    * Books that are checked out cannot be edited until they are checked by in
* Remove a book from being available in the library
* Check book out of library
* Check book in to library
* View status of all books check in to and out of library
* Logs all actions taken inside of the library application to an event store

## Improvements

The project CRUD interface for managin books in a library, and an event store for tracking actions taking in the system.

* Read book(s) state from reduction of event store rather than mutating book information in the database.
* Add server side validation of book fields when adding book
* Add pagination and sorting functionality to allow sorting by book by attributes (e.g. title, author, etc.).
* Auto suggest known ISBN's when typing in corresponding field and dynamically populate book details based on [ISB API](http://www.isbndb.com/").
* Add more content to books (e.g. cover image, publish date, etc.)
* As more routes are added to the application, the templating logic could be abstracted into a dedicated package that would also parse an option `templates/_components` directory to allow for nesting reusable component partials into route pages.
* Add a `templates/_components` directory to encapsulate component markup.
* Add a `scripts/components` directory encapselate component scripts in a single class.
