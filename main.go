package main

import (
	"fmt"
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
	fmt.Fprintf(w, "<h1>%s</h1>", p.Title)
	fmt.Fprintf(w, "<div>%s</div>", p.Body)
}

func listHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "<h1>Index</h1>")
	entries, err := os.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintln(w, "<ul>")
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".txt") {
			title := strings.TrimSuffix(e.Name(), ".txt")
			fmt.Fprintf(w, "<li><a href=\"view/%s\">%s</a></li>", title, title)
		}
	}
	fmt.Fprintln(w, "</ul>")
}

func main() {
	http.HandleFunc("/", listHandler)
	http.HandleFunc("/view/", viewHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
