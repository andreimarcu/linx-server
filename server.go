package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"regexp"
	"os"
	"fmt"

	"github.com/flosch/pongo2"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web/middleware"
)

var Config struct {
	bind     string
	filesDir string
	noLogs   bool
	siteName string
	siteURL  string
}

func main() {
	flag.StringVar(&Config.bind, "b", "127.0.0.1:8080",
		"host to bind to (default: 127.0.0.1:8080)")
	flag.StringVar(&Config.filesDir, "filespath", "files/",
		"path to files directory (including trailing slash)")
	flag.BoolVar(&Config.noLogs, "nologs", false,
		"remove stdout output for each request")
	flag.StringVar(&Config.siteName, "sitename", "linx",
		"name of the site")
	flag.StringVar(&Config.siteURL, "siteurl", "http://"+Config.bind+"/",
		"site base url (including trailing slash)")
	flag.Parse()

	if Config.noLogs {
		goji.Abandon(middleware.Logger)
	}

	// make directory if needed
	err := os.MkdirAll(Config.filesDir, 0755)
	if err != nil {
		fmt.Printf("Error: could not create files directory")
		os.exit(1)
	}

	// Template Globals
	pongo2.DefaultSet.Globals["sitename"] = Config.siteName

	// Routing setup
	nameRe := regexp.MustCompile(`^/(?P<name>[a-z0-9-\.]+)$`)
	selifRe := regexp.MustCompile(`^/selif/(?P<name>[a-z0-9-\.]+)$`)

	goji.Get("/", indexHandler)

	goji.Post("/upload", uploadPostHandler)
	goji.Post("/upload/", http.RedirectHandler("/upload", 301))
	goji.Put("/upload", uploadPutHandler)
	goji.Put("/upload/:name", uploadPutHandler)

	goji.Get("/static/*", http.StripPrefix("/static/",
		http.FileServer(http.Dir("static/"))))
	goji.Get(nameRe, fileDisplayHandler)
	goji.Get(selifRe, fileServeHandler)
	goji.NotFound(notFoundHandler)

	listener, err := net.Listen("tcp", Config.bind)
	if err != nil {
		log.Fatal("Could not bind: ", err)
	}

	goji.ServeListener(listener)
}
