package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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
func ServeFile(w http.ResponseWriter, r *http.Request, path string) {
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

	w.Write(render.Render(dat))
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
func (MainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	Custom404(w, r)
}

type MainHandler struct{}

func main() {
	// http.Handle("/metrics", promhttp.Handler())
	addr := flag.String("addr", "0.0.0.0", "address to serve on")
	port := flag.Int("port", 3000, "port to serve on")

	metricsAddr := flag.String("metricsAddr", "0.0.0.0", "address to serve on")
	metricsPort := flag.Int("metricsPort", 8080, "port to serve on")

	go func() {
		err := http.ListenAndServe(fmt.Sprintf("%s:%d", *addr, *port), MainHandler{})
		if err != nil {
			log.Fatal(err)
		}
	}()

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", *metricsAddr, *metricsPort), promhttp.Handler())
	if err != nil {
		log.Fatal(err)
	}
}
