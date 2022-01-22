package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	return os.WriteFile(p.Title+".txt", p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	body, err := os.ReadFile(title + ".txt")
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	var titles []string
	entries, err := os.ReadDir(".")
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
	p, err := loadPage(title)
	if os.IsNotExist(err) {
		http.Error(w, "page does not exist", http.StatusBadRequest)
		return
	}
	if err != nil {
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
	if err := templ.Execute(w, struct {
		Title string
		Body  template.HTML
	}{
		Title: p.Title,
		Body:  template.HTML(p.Body),
	}); err != nil {
		log.Printf("could not execute template: %v", err)
	}
}

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimPrefix(r.URL.Path, "/edit/")
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
	if err := templ.Execute(w, struct {
		Title string
		Body  template.HTML
	}{
		Title: p.Title,
		Body:  template.HTML(p.Body),
	}); err != nil {
		log.Printf("could not execute template: %v", err)
	}
}

func saveHandler(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Title: strings.TrimPrefix(r.URL.Path, "/save/"),
		Body:  []byte(r.FormValue("body")),
	}
	if err := p.save(); err != nil {
		log.Printf("could not save page: %v", err)
		http.Error(w, "could not save page", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+p.Title, http.StatusFound)
}

func main() {
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	http.HandleFunc("/save/", saveHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
