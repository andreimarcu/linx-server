package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"regexp"

	"github.com/flosch/pongo2"
	"github.com/zenazn/goji"
)

var Config struct {
	bind     string
	filesDir string
	siteName string
}

func main() {
	flag.StringVar(&Config.bind, "b", "127.0.0.1:8080",
		"host to bind to (default: 127.0.0.1:8080)")
	flag.StringVar(&Config.filesDir, "d", "files/",
		"path to files directory (default: files/)")
	flag.StringVar(&Config.siteName, "n", "linx",
		"name of the site")
	flag.Parse()

	// Disable template caching -- keep until out of pre-alpha
	pongo2.DefaultSet.Debug = true // will keep this until out of pre-alpha

	// Template Globals
	pongo2.DefaultSet.Globals["sitename"] = Config.siteName

	// Routing setup
	nameRe := regexp.MustCompile(`^/(?P<name>[a-z0-9-\.]+)$`)
	selifRe := regexp.MustCompile(`^/selif/(?P<name>[a-z0-9-\.]+)$`)

	goji.Get("/", indexHandler)
	goji.Post("/upload", uploadPostHandler)
	goji.Put("/upload", uploadPutHandler)
	goji.Get("/static/*", http.StripPrefix("/static/",
		http.FileServer(http.Dir("static/"))))
	goji.Get(nameRe, fileDisplayHandler)
	goji.Get(selifRe, fileServeHandler)

	listener, err := net.Listen("tcp", Config.bind)
	if err != nil {
		log.Fatal("Could not bind: ", err)
	}

	goji.ServeListener(listener)
}
