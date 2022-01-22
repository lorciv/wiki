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
	fmt.Fprintf(w, "<ul><li><a href=\"/edit/%s\">Edit</a></li><li><a href=\"/list\">Index</a></li></ul>", title)
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
	fmt.Fprintf(w, "<h1>Edit: %s</h1>", p.Title)
	fmt.Fprintf(w, "<form action=\"/save/%s\" method=\"post\">", p.Title)
	fmt.Fprintf(w, "<div><textarea name=\"body\" rows=\"30\" cols=\"100\">%s</textarea></div>", p.Body)
	fmt.Fprintf(w, "<div><input type=\"submit\"></div>")
	fmt.Fprintf(w, "</form>")
}

func main() {
	http.HandleFunc("/list", listHandler)
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/edit/", editHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
