package main

import (
	"fmt"
	"log"
	"os"
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

func main() {
	p := Page{
		Title: "Babbo Natale",
		Body:  []byte("Babbo Natale è un personaggio di tanti racconti per bambini"),
	}
	if err := p.save(); err != nil {
		log.Fatalf("could not save page: %v", err)
	}

	p2, err := loadPage("Babbo Natale")
	if err != nil {
		log.Fatalf("could not load page: %v", err)
	}
	fmt.Println(p2.Title, string(p2.Body))
}
