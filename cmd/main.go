package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/ubarar/jekill/render"
)

// Serve an existing file to the response writer
// If the requested file is markdown, render it.
func ServeFile(w http.ResponseWriter, r *http.Request, path string) {
	if !strings.HasSuffix(path, ".md") {
		http.ServeFile(w, r, path)
		return
	}

	// .md file, it must be rendered
	dat, err := os.ReadFile(path)
	if err != nil {
		Custom500(w)
		return
	}

	w.Write(render.Render(dat))
	w.Header().Add("Content-Type", "text/html")
}

func Custom404(w http.ResponseWriter) {
	w.WriteHeader(404)
	w.Write([]byte("Could not find"))
}

func Custom500(w http.ResponseWriter) {
	w.WriteHeader(500)
	w.Write([]byte("Internal error detected"))
}

// Serve files from the current directory. Special rules for
// the index page.
// If a file doesn't exist, see if there's a corresponding markdown file. This is
// for cases where the user wants /about, but the page is about.md or about.html
func MainHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.RequestURI(), "/")

	if path == "" {
		path = "index"
	}

	if _, err := os.Stat(path); err == nil {
		ServeFile(w, r, path)
		return
	}

	if _, err := os.Stat(path + ".md"); err == nil {
		ServeFile(w, r, path+".md")
		return
	}

	if _, err := os.Stat(path + ".html"); err == nil {
		ServeFile(w, r, path+".html")
		return
	}

	Custom404(w)
}

func main() {
	addr := flag.String("addr", "0.0.0.0", "address to serve on")
	port := flag.Int("port", 3000, "port to serve on")
	http.HandleFunc("/", MainHandler)

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", *addr, *port), nil)
	if err != nil {
		log.Fatal(err)
	}
}
