package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"

	"github.com/flosch/pongo2"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web/middleware"
)

var Config struct {
	bind     string
	filesDir string
	metaDir  string
	noLogs   bool
	siteName string
	siteURL  string
}

func main() {
	flag.StringVar(&Config.bind, "b", "127.0.0.1:8080",
		"host to bind to (default: 127.0.0.1:8080)")
	flag.StringVar(&Config.filesDir, "filespath", "files/",
		"path to files directory")
	flag.StringVar(&Config.metaDir, "metapath", "meta/",
		"path to metadata directory")
	flag.BoolVar(&Config.noLogs, "nologs", false,
		"remove stdout output for each request")
	flag.StringVar(&Config.siteName, "sitename", "linx",
		"name of the site")
	flag.StringVar(&Config.siteURL, "siteurl", "http://"+Config.bind+"/",
		"site base url")
	flag.Parse()

	if Config.noLogs {
		goji.Abandon(middleware.Logger)
	}

	// make directories if needed
	var err error

	err = os.MkdirAll(Config.filesDir, 0755)
	if err != nil {
		fmt.Printf("Error: could not create files directory\n")
		os.Exit(1)
	}

	err = os.MkdirAll(Config.metaDir, 0700)
	if err != nil {
		fmt.Printf("Error: could not create metadata directory\n")
		os.Exit(1)
	}

	// ensure siteURL ends wth '/'
	if lastChar := Config.siteURL[len(Config.siteURL)-1:]; lastChar != "/" {
		Config.siteURL = Config.siteURL + "/"
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
