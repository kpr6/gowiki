package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
)

type Page struct {
	Title string
	Body  []byte
	Index []string
}

func (p *Page) save() error {
	if _, err := os.Stat("./data"); os.IsNotExist(err) {
		_ = os.Mkdir("./data", 0660)
	}
	filename := filepath.Join("./data", p.Title+".txt")
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := filepath.Join("./data", title+".txt")
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
	if title == "" {
		title = "FrontPage"
	}
	p, err := loadPage(title)
	if err != nil {
		http.Redirect(w, r, "/edit/"+title, http.StatusFound)
		return
	}
	files, err := ioutil.ReadDir("./data")
	if err != nil {
		panic(err)
	}
	p.Index = []string{}
	for _, f := range files {
		name := f.Name()
		l := len(name) - len(filepath.Ext(name))
		p.Index = append(p.Index, name[:l])
	}
	fmt.Println(p.Index)
	renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
	body := r.FormValue("body")
	p := &Page{Title: title, Body: []byte(body)}
	err := p.save()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

var templates = template.Must(template.ParseGlob("./tmpl/*"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
	err := templates.ExecuteTemplate(w, tmpl+".html", p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

var validPath = regexp.MustCompile(`^/(edit|save|view)/([a-zA-Z0-9]+)$`)

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m := validPath.FindStringSubmatch(r.URL.Path)
		if m == nil {
			http.NotFound(w, r)
			return
		}
		fn(w, r, m[2])
	}
}

func main() {
	http.HandleFunc("/view/", makeHandler(viewHandler))
	http.HandleFunc("/edit/", makeHandler(editHandler))
	http.HandleFunc("/save/", makeHandler(saveHandler))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.Redirect(w, r, "/view/FrontPage", http.StatusFound) })

	log.Fatal(http.ListenAndServe(":8080", nil))
}
