package main

import (
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Page struct {
	Title   string
	RawBody []byte
	Body    []Element // available only after call to parse
}

func (p *Page) parse() error {
	l := lexer{
		input: string(p.RawBody),
	}
	state := lex
	for state != nil {
		state = state(&l)
	}
	fmt.Println(l.tokens)
	par := parser{
		tokens: l.tokens,
	}
	var err error
	p.Body, err = par.parse()
	return err
}

func (p *Page) save() error {
	return os.WriteFile("pages/"+p.Title+".txt", p.RawBody, 0600)
}

func loadPage(title string) (*Page, error) {
	body, err := os.ReadFile("pages/" + title + ".txt")
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, RawBody: body}, nil
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	var titles []string
	entries, err := os.ReadDir("pages")
	if err != nil {
		log.Fatal(err)
	}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			titles = append(titles, strings.TrimSuffix(e.Name(), ".txt"))
		}
	}

	// TODO: template caching
	templ, err := template.ParseFiles("templates/list.html")
	if err != nil {
		log.Printf("could not parse template: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if err := templ.Execute(w, titles); err != nil {
		log.Printf("could not execute template: %v", err)
	}
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimPrefix(r.URL.Path, "/view/")
	if !validTitle.MatchString(title) {
		http.Error(w, "invalid page title", http.StatusBadRequest)
		return
	}
	p, err := loadPage(title)
	if os.IsNotExist(err) {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	if err := p.parse(); err != nil {
		log.Printf("could not parse page: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// TODO: template caching
	templ, err := template.ParseFiles("templates/view.html")
	if err != nil {
		log.Printf("could not parse template: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if err := templ.Execute(w, p); err != nil {
		log.Printf("could not execute template: %v", err)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimPrefix(r.URL.Path, "/edit/")
	if !validTitle.MatchString(title) {
		http.Error(w, "invalid page title", http.StatusBadRequest)
		return
	}
	p, err := loadPage(title)
	if os.IsNotExist(err) {
		p = &Page{Title: title}
	} else if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// TODO: template caching
	templ, err := template.ParseFiles("templates/edit.html")
	if err != nil {
		log.Printf("could not parse template: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	if err := templ.Execute(w, p); err != nil {
		log.Printf("could not execute template: %v", err)
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Title:   strings.TrimPrefix(r.URL.Path, "/save/"),
		RawBody: []byte(r.FormValue("body")),
	}
	if !validTitle.MatchString(p.Title) {
		http.Error(w, "invalid page title", http.StatusBadRequest)
		return
	}
	if err := p.save(); err != nil {
		log.Printf("could not save page: %v", err)
		http.Error(w, "could not save page", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+p.Title, http.StatusFound)
}

var validTitle = regexp.MustCompile("^([a-zA-Z0-9]+)$")

var port = flag.Int("p", 8080, "port")

func main() {
	flag.Parse()

	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)

	addr := fmt.Sprintf(":%d", *port)
	log.Fatal(http.ListenAndServe(addr, nil))
}
