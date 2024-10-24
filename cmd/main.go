package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"path/filepath"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ubarar/jekill/render"
)

var (
	requestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "jekill_requests_total",
		Help: "The total number of requests",
	},
		[]string{"path", "code"})
)

// Serve an existing file to the response writer
// If the requested file is markdown, render it.
func (s Service) ServeFile(w http.ResponseWriter, r *http.Request, path string) {
	requestsTotal.WithLabelValues(r.URL.RequestURI(), "200").Inc()
	if !strings.HasSuffix(path, ".md") {
		http.ServeFile(w, r, path)
		return
	}

	// .md file, it must be rendered
	dat, err := os.ReadFile(path)
	if err != nil {
		Custom500(w, r)
		return
	}

	w.Write(s.Renderer.Render(dat))
	w.Header().Add("Content-Type", "text/html")
}

func Custom404(w http.ResponseWriter, r *http.Request) {
	requestsTotal.WithLabelValues(r.URL.RequestURI(), "404").Inc()
	w.WriteHeader(404)
	w.Write([]byte("Could not find"))
}

func Custom500(w http.ResponseWriter, r *http.Request) {
	requestsTotal.WithLabelValues(r.URL.RequestURI(), "404").Inc()
	w.WriteHeader(500)
	w.Write([]byte("Internal error detected"))
}

// Serve files from the current directory. Special rules for
// the index page.
// If a file doesn't exist, see if there's a corresponding markdown file. This is
// for cases where the user wants /about, but the page is about.md or about.html
func (s Service) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.RequestURI(), "/")
	if path == "" {
		path = "index"
	}

	// where the file actually lives on the filesystem
	diskPath := filepath.Join(s.Path, path)

	fmt.Println(diskPath)

	if info, err := os.Stat(diskPath); err == nil && !info.IsDir() {
		s.ServeFile(w, r, diskPath)
		return
	}

	if info, err := os.Stat(diskPath + ".md"); err == nil && !info.IsDir() {
		s.ServeFile(w, r, diskPath+".md")
		return
	}

	if info, err := os.Stat(diskPath + ".html"); err == nil && !info.IsDir() {
		s.ServeFile(w, r, diskPath+".html")
		return
	}

	Custom404(w, r)
}

type Service struct{
	Path string
	Renderer *render.Renderer
}

func main() {
	path := flag.String("path", "", "path where we should look for all files")
	addr := flag.String("addr", "0.0.0.0", "address to serve on")
	port := flag.Int("port", 3000, "port to serve on")

	flag.Parse()

	metricsAddr := flag.String("metricsAddr", "0.0.0.0", "address to serve on")
	metricsPort := flag.Int("metricsPort", 8080, "port to serve on")

	go func() {
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", *addr, *port), Service{Path: *path, Renderer: render.NewRenderer(*path)})
		if err != nil {
			log.Fatal(err)
		}
	}()

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", *metricsAddr, *metricsPort), promhttp.Handler())
	if err != nil {
		log.Fatal(err)
	}
}
