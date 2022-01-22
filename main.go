package main

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Page struct {
	Title string
	Body  []string
}

func (p *Page) save() error {
	f, err := os.Create("pages/" + p.Title + ".txt")
	if err != nil {
		return err
	}
	defer f.Close()

	for _, s := range p.Body {
		fmt.Fprintln(f, s)
	}
	return nil
}

func loadPage(title string) (*Page, error) {
	f, err := os.Open("pages/" + title + ".txt")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	p := &Page{Title: title}
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		line := scan.Text()
		if line == "" {
			continue
		}
		p.Body = append(p.Body, line)
	}
	if err := scan.Err(); err != nil {
		return nil, err
	}
	return p, nil
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

// func saveHandler(w http.ResponseWriter, r *http.Request) {
// 	p := &Page{
// 		Title: strings.TrimPrefix(r.URL.Path, "/save/"),
// 		Body:  []byte(r.FormValue("body")),
// 	}
// 	if !validTitle.MatchString(p.Title) {
// 		http.Error(w, "invalid page title", http.StatusBadRequest)
// 		return
// 	}
// 	if err := p.save(); err != nil {
// 		log.Printf("could not save page: %v", err)
// 		http.Error(w, "could not save page", http.StatusInternalServerError)
// 		return
// 	}
// 	http.Redirect(w, r, "/view/"+p.Title, http.StatusFound)
// }

var validTitle = regexp.MustCompile("^([a-zA-Z0-9]+)$")

// func main() {
// 	http.HandleFunc("/list", listHandler)
// 	http.HandleFunc("/view/", viewHandler)
// 	http.HandleFunc("/edit/", editHandler)
// 	http.HandleFunc("/save/", saveHandler)
// 	log.Fatal(http.ListenAndServe(":8080", nil))
// }
