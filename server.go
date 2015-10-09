package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/flosch/pongo2"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web/middleware"
)

var Config struct {
	bind                      string
	filesDir                  string
	metaDir                   string
	siteName                  string
	siteURL                   string
	certFile                  string
	keyFile                   string
	contentSecurityPolicy     string
	fileContentSecurityPolicy string
	xFrameOptions             string
	maxSize                   int64
	noLogs                    bool
	allowHotlink              bool
	fastcgi                   bool
	remoteUploads             bool
}

var Templates = make(map[string]*pongo2.Template)
var TemplateSet *pongo2.TemplateSet
var staticBox *rice.Box
var timeStarted time.Time
var timeStartedStr string

func setup() {
	goji.Use(ContentSecurityPolicy(CSPOptions{
		policy: Config.contentSecurityPolicy,
		frame:  Config.xFrameOptions,
	}))

	if Config.noLogs {
		goji.Abandon(middleware.Logger)
	}

	// make directories if needed
	err := os.MkdirAll(Config.filesDir, 0755)
	if err != nil {
		log.Fatal("Could not create files directory:", err)
	}

	err = os.MkdirAll(Config.metaDir, 0700)
	if err != nil {
		log.Fatal("Could not create metadata directory:", err)
	}

	// ensure siteURL ends wth '/'
	if lastChar := Config.siteURL[len(Config.siteURL)-1:]; lastChar != "/" {
		Config.siteURL = Config.siteURL + "/"
	}

	// Template setup
	p2l, err := NewPongo2TemplatesLoader()
	if err != nil {
		log.Fatal("Error: could not load templates", err)
	}
	TemplateSet := pongo2.NewSet("templates", p2l)
	TemplateSet.Globals["sitename"] = Config.siteName
	TemplateSet.Globals["siteurl"] = Config.siteURL
	err = populateTemplatesMap(TemplateSet, Templates)
	if err != nil {
		log.Fatal("Error: could not load templates", err)
	}

	staticBox = rice.MustFindBox("static")
	timeStarted = time.Now()
	timeStartedStr = strconv.FormatInt(timeStarted.Unix(), 10)

	// Routing setup
	nameRe := regexp.MustCompile(`^/(?P<name>[a-z0-9-\.]+)$`)
	selifRe := regexp.MustCompile(`^/selif/(?P<name>[a-z0-9-\.]+)$`)
	selifIndexRe := regexp.MustCompile(`^/selif/$`)
	torrentRe := regexp.MustCompile(`^/(?P<name>[a-z0-9-\.]+)/torrent$`)

	goji.Get("/", indexHandler)
	goji.Get("/paste/", pasteHandler)
	goji.Get("/paste", http.RedirectHandler("/paste/", 301))
	goji.Get("/API/", apiDocHandler)
	goji.Get("/API", http.RedirectHandler("/API/", 301))

	if Config.remoteUploads {
		goji.Get("/upload", uploadRemote)
		goji.Get("/upload/", uploadRemote)
	}

	goji.Post("/upload", uploadPostHandler)
	goji.Post("/upload/", uploadPostHandler)
	goji.Put("/upload", uploadPutHandler)
	goji.Put("/upload/:name", uploadPutHandler)
	goji.Delete("/:name", deleteHandler)

	goji.Get("/static/*", staticHandler)
	goji.Get("/favicon.ico", staticHandler)
	goji.Get("/robots.txt", staticHandler)
	goji.Get(nameRe, fileDisplayHandler)
	goji.Get(selifRe, fileServeHandler)
	goji.Get(selifIndexRe, unauthorizedHandler)
	goji.Get(torrentRe, fileTorrentHandler)
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
	flag.BoolVar(&Config.allowHotlink, "allowhotlink", false,
		"Allow hotlinking of files")
	flag.StringVar(&Config.siteName, "sitename", "linx",
		"name of the site")
	flag.StringVar(&Config.siteURL, "siteurl", "http://"+Config.bind+"/",
		"site base url (including trailing slash)")
	flag.Int64Var(&Config.maxSize, "maxsize", 4*1024*1024*1024,
		"maximum upload file size in bytes (default 4GB)")
	flag.StringVar(&Config.certFile, "certfile", "",
		"path to ssl certificate (for https)")
	flag.StringVar(&Config.keyFile, "keyfile", "",
		"path to ssl key (for https)")
	flag.BoolVar(&Config.fastcgi, "fastcgi", false,
		"serve through fastcgi")
	flag.BoolVar(&Config.remoteUploads, "remoteuploads", false,
		"enable remote uploads")
	flag.StringVar(&Config.contentSecurityPolicy, "contentsecuritypolicy",
		"default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; referrer none;",
		"value of default Content-Security-Policy header")
	flag.StringVar(&Config.fileContentSecurityPolicy, "filecontentsecuritypolicy",
		"default-src 'none'; img-src 'self'; object-src 'self'; media-src 'self'; sandbox; referrer none;",
		"value of Content-Security-Policy header for file access")
	flag.StringVar(&Config.xFrameOptions, "xframeoptions", "SAMEORIGIN",
		"value of X-Frame-Options header")
	flag.Parse()

	setup()

	if Config.fastcgi {
		listener, err := net.Listen("tcp", Config.bind)
		if err != nil {
			log.Fatal("Could not bind: ", err)
		}

		log.Printf("Serving over fastcgi, bound on %s, using siteurl %s", Config.bind, Config.siteURL)
		fcgi.Serve(listener, goji.DefaultMux)
	} else if Config.certFile != "" {
		log.Printf("Serving over https, bound on %s, using siteurl %s", Config.bind, Config.siteURL)
		err := graceful.ListenAndServeTLS(Config.bind, Config.certFile, Config.keyFile, goji.DefaultMux)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("Serving over http, bound on %s, using siteurl %s", Config.bind, Config.siteURL)
		err := graceful.ListenAndServe(Config.bind, goji.DefaultMux)
		if err != nil {
			log.Fatal(err)
		}
	}
}
