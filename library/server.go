package library

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/rs/xid"
)

type Server struct {
	BookRepo  BookRepo
	EventRepo EventRepo
	Router    *mux.Router
	Env       map[string]string
}

func (s *Server) Routes() {
	// Define router, static files, and middleware
	s.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	// Routes to handle requests from browsers and HTML forms
	s.Router.HandleFunc("/", s.handleIndex()).Methods("GET")
	s.Router.HandleFunc("/books", s.handleBooks()).Methods("GET")
	s.Router.HandleFunc("/books/add", s.handleAddBook()).Methods("GET", "POST")
	s.Router.HandleFunc("/books/remove", s.handleRemoveBook()).Methods("POST")
	s.Router.HandleFunc("/books/check-in", s.handleBookCheckIn()).Methods("POST")
	s.Router.HandleFunc("/books/check-out", s.handleBookCheckOut()).Methods("POST")
	s.Router.HandleFunc("/books/{id:[0-9]+}", s.handleViewBook()).Methods("GET", "POST")
	s.Router.HandleFunc("/books/report", s.handleReportBooks()).Methods("GET")
}

/**
 * HTTP handler for get requests to view the library homepage
 */
func (s *Server) handleIndex() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		name := "index"
		title := "Library"

		// Handle 404 routes
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			name = "404"
			title = "Uh oh"
		}

		s.getTemplate(name, nil).Execute(w, map[string]interface{}{
			"title": title,
			"env":   s.Env,
		})
	}
}

/**
 * HTTP handler for get requests to view all books
 */
func (s *Server) handleBooks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		books, err := s.BookRepo.All()
		if err != nil {
			fmt.Println("Unable to get books")
		}

		s.getTemplate("books/index", nil).Execute(w, map[string]interface{}{
			"title":          "Add Book",
			"env":            s.Env,
			csrf.TemplateTag: csrf.TemplateField(r),
			"books":          books,
		})
	}
}

/**
 * HTTP handler for get/post requests to add a book to the stacks
 */
func (s *Server) handleAddBook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Handle save book on post
		if r.Method == http.MethodPost {
			// Create new book from form values
			b, err := s.BookRepo.New(
				r.FormValue("title"),
				r.FormValue("author"),
				r.FormValue("isbn"),
				r.FormValue("description"),
			)
			if err != nil {
				fmt.Printf("Cannot create book: %s", err)
				http.Redirect(w, r, "/books", http.StatusSeeOther)
			}

			if err = s.EventRepo.BookAdded(b.ID); err != nil {
				// TODO: Handle event error
			}

			http.Redirect(w, r, "/books", http.StatusSeeOther)
		}

		// Handle save book form on get
		s.getTemplate("books/add", nil).Execute(w, map[string]interface{}{
			"title":          "Add Book",
			"env":            s.Env,
			csrf.TemplateTag: csrf.TemplateField(r),
		})
	}
}

/**
 * HTTP handler for post requests to remove a book from the stacks
 */
func (s *Server) handleRemoveBook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			fmt.Printf("Cannot get ID from form: %s", err)
			http.Redirect(w, r, "/books", http.StatusSeeOther)
		}

		err = s.BookRepo.Delete(id)
		if err != nil {
			fmt.Printf("Cannot delete book with ID %d: %s", id, err)
			http.Redirect(w, r, "/books", http.StatusSeeOther)
		}

		if err = s.EventRepo.BookRemoved(id); err != nil {
			// TODO: Handle event error
		}

		http.Redirect(w, r, "/books", http.StatusSeeOther)
	}
}

/**
 * HTTP handler for post requests to check in a book
 */
func (s *Server) handleBookCheckIn() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			http.Redirect(w, r, "/books", http.StatusSeeOther)
		}

		if err = s.BookRepo.CheckIn(id); err != nil {
			fmt.Printf("Unable to check in book %d: %s\n", id, err)
		}

		if err = s.EventRepo.BookCheckedIn(id); err != nil {
			// TODO: Handle event error
		}

		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}

/**
 * HTTP handler for post requests to check out a book
 */
func (s *Server) handleBookCheckOut() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.FormValue("id"))
		if err != nil {
			http.Redirect(w, r, "/books", http.StatusSeeOther)
		}

		if err = s.BookRepo.CheckOut(id); err != nil {
			fmt.Printf("Unable to check out book %d: %s\n", id, err)
		}

		if err = s.EventRepo.BookCheckedOut(id); err != nil {
			// TODO: Handle event error
		}

		http.Redirect(w, r, r.Referer(), http.StatusSeeOther)
	}
}

/**
 * HTTP handler for get/post requests to view a book
 */
func (s *Server) handleViewBook() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		id, err := strconv.Atoi(vars["id"])
		if err != nil {
			http.Redirect(w, r, "/books", http.StatusMovedPermanently)
		}

		b, err := s.BookRepo.One(id)
		if err != nil {
			fmt.Printf("Cannot find book %d\n", id)
			http.Redirect(w, r, "/books", http.StatusSeeOther)
		}

		es, err := s.EventRepo.AllForBook(id)
		if err != nil {
			fmt.Printf("Cannot find book %d\n", id)
			http.Redirect(w, r, "/books", http.StatusSeeOther)
		}

		// Do not allow checkout out books to be edited
		if b.IsCheckedOut() {
			http.Redirect(w, r, "/books", http.StatusGone)
		}

		// Handle save book on post
		if r.Method == http.MethodPost {
			// Update book from form values
			b.Title = r.FormValue("title")
			b.Author = r.FormValue("author")
			b.ISBN = r.FormValue("isbn")
			b.Description = r.FormValue("description")
			if err = s.BookRepo.Update(b); err != nil {
				fmt.Printf("Cannot update book: %s", err)
				http.Redirect(w, r, "/books", http.StatusSeeOther)
			}

			if err = s.EventRepo.BookAdded(b.ID); err != nil {
				// TODO: Handle event error
			}

			http.Redirect(w, r, fmt.Sprintf(r.Referer(), id), http.StatusSeeOther)
		}

		// Handle save book form on get
		s.getTemplate("books/_book", nil).Execute(w, map[string]interface{}{
			"title":          "Add Book",
			"env":            s.Env,
			csrf.TemplateTag: csrf.TemplateField(r),
			"book":           b,
			"events":         es,
		})
	}
}

/**
 * HTTP handler for get requests to view status report of books
 */
func (s *Server) handleReportBooks() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		books, err := s.BookRepo.All()
		if err != nil {
			fmt.Println("Unable to get books")
		}

		var booksIn []*book
		var booksOut []*book

		// TODO: Might be a better way than looping over books after repo call
		for _, b := range books {
			if b.IsCheckedOut() {
				booksOut = append(booksOut, b)
			} else {
				booksIn = append(booksIn, b)
			}
		}

		s.getTemplate("books/report", nil).Execute(w, map[string]interface{}{
			"title":          "Add Book",
			"env":            s.Env,
			csrf.TemplateTag: csrf.TemplateField(r),
			"booksIn":        booksIn,
			"booksOut":       booksOut,
		})
	}
}

/**
 * Helper function to load all required layouts and partials to load template
 */
func (s *Server) getTemplate(name string, fm template.FuncMap) *template.Template {
	funcMap := template.FuncMap{
		"now": func() int {
			return time.Now().Year()
		},
		"uniqueID": func() string {
			return xid.New().String()
		},
	}
	// Merge custom funcMap
	for k, v := range fm {
		funcMap[k] = v
	}

	t, err := template.New("main.html").Funcs(funcMap).ParseFiles(
		"templates/_layouts/main.html",
		"templates/_meta/data.html",
		"templates/_meta/favicons.html",
		fmt.Sprintf("templates/%s.html", name),
	)
	if err != nil {
		fmt.Printf("Unable to load template %s: \n", name, err)
	}

	return t
}
