package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"

	"github.com/GeertJohan/go.rice"
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

var Templates = make(map[string]*pongo2.Template)
var TemplateSet *pongo2.TemplateSet

func setup() {
	if Config.noLogs {
		goji.Abandon(middleware.Logger)
	}

	// make directories if needed
	err := os.MkdirAll(Config.filesDir, 0755)
	if err != nil {
		fmt.Println("Error: could not create files directory")
		os.Exit(1)
	}

	err = os.MkdirAll(Config.metaDir, 0700)
	if err != nil {
		fmt.Println("Error: could not create metadata directory")
		os.Exit(1)
	}

	// ensure siteURL ends wth '/'
	if lastChar := Config.siteURL[len(Config.siteURL)-1:]; lastChar != "/" {
		Config.siteURL = Config.siteURL + "/"
	}

	// Template setup
	p2l, err := NewPongo2Loader("templates")
	if err != nil {
		fmt.Println("Error: could not load templates")
		os.Exit(1)
	}
	TemplateSet := pongo2.NewSet("templates", p2l)
	TemplateSet.Globals["sitename"] = Config.siteName
	err = populateTemplatesMap(TemplateSet, Templates)
	if err != nil {
		fmt.Println("Error: could not load templates")
		os.Exit(1)
	}

	// Routing setup
	nameRe := regexp.MustCompile(`^/(?P<name>[a-z0-9-\.]+)$`)
	selifRe := regexp.MustCompile(`^/selif/(?P<name>[a-z0-9-\.]+)$`)

	goji.Get("/", indexHandler)

	goji.Post("/upload", uploadPostHandler)
	goji.Post("/upload/", uploadPostHandler)
	goji.Put("/upload", uploadPutHandler)
	goji.Put("/upload/:name", uploadPutHandler)

	staticBox := rice.MustFindBox("static")
	goji.Get("/static/*", http.StripPrefix("/static/",
		http.FileServer(staticBox.HTTPBox())))
	goji.Get(nameRe, fileDisplayHandler)
	goji.Get(selifRe, fileServeHandler)
	goji.NotFound(notFoundHandler)
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
		"site base url (including trailing slash)")
	flag.Parse()

	setup()

	listener, err := net.Listen("tcp", Config.bind)
	if err != nil {
		log.Fatal("Could not bind: ", err)
	}

	goji.ServeListener(listener)
}
