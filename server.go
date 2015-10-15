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
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
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
	realIp                    bool
	noLogs                    bool
	allowHotlink              bool
	fastcgi                   bool
	remoteUploads             bool
	authFile                  string
	remoteAuthFile            string
}

var Templates = make(map[string]*pongo2.Template)
var TemplateSet *pongo2.TemplateSet
var staticBox *rice.Box
var timeStarted time.Time
var timeStartedStr string
var remoteAuthKeys []string

func setup() *web.Mux {
	mux := web.New()

	// middleware
	mux.Use(middleware.RequestID)

	if Config.realIp {
		mux.Use(middleware.RealIP)
	}

	if !Config.noLogs {
		mux.Use(middleware.Logger)
	}

	mux.Use(middleware.Recoverer)
	mux.Use(middleware.AutomaticOptions)
	mux.Use(ContentSecurityPolicy(CSPOptions{
		policy: Config.contentSecurityPolicy,
		frame:  Config.xFrameOptions,
	}))

	if Config.authFile != "" {
		mux.Use(UploadAuth(AuthOptions{
			AuthFile:      Config.authFile,
			UnauthMethods: []string{"GET", "HEAD", "OPTIONS", "TRACE"},
		}))
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
	TemplateSet.Globals["using_auth"] = Config.authFile != ""
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

	if Config.authFile == "" {
		mux.Get("/", indexHandler)
		mux.Get("/paste/", pasteHandler)
	} else {
		mux.Get("/", http.RedirectHandler("/API", 303))
		mux.Get("/paste/", http.RedirectHandler("/API/", 303))
	}
	mux.Get("/paste", http.RedirectHandler("/paste/", 301))

	mux.Get("/API/", apiDocHandler)
	mux.Get("/API", http.RedirectHandler("/API/", 301))

	if Config.remoteUploads {
		mux.Get("/upload", uploadRemote)
		mux.Get("/upload/", uploadRemote)

		if Config.remoteAuthFile != "" {
			remoteAuthKeys = readAuthKeys(Config.remoteAuthFile)
		}
	}

	mux.Post("/upload", uploadPostHandler)
	mux.Post("/upload/", uploadPostHandler)
	mux.Put("/upload", uploadPutHandler)
	mux.Put("/upload/", uploadPutHandler)
	mux.Put("/upload/:name", uploadPutHandler)

	mux.Delete("/:name", deleteHandler)

	mux.Get("/static/*", staticHandler)
	mux.Get("/favicon.ico", staticHandler)
	mux.Get("/robots.txt", staticHandler)
	mux.Get(nameRe, fileDisplayHandler)
	mux.Get(selifRe, fileServeHandler)
	mux.Get(selifIndexRe, unauthorizedHandler)
	mux.Get(torrentRe, fileTorrentHandler)
	mux.NotFound(notFoundHandler)

	return mux
}

func main() {
	flag.StringVar(&Config.bind, "bind", "127.0.0.1:8080",
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
	flag.BoolVar(&Config.realIp, "realip", false,
		"use X-Real-IP/X-Forwarded-For headers as original host")
	flag.BoolVar(&Config.fastcgi, "fastcgi", false,
		"serve through fastcgi")
	flag.BoolVar(&Config.remoteUploads, "remoteuploads", false,
		"enable remote uploads")
	flag.StringVar(&Config.authFile, "authfile", "",
		"path to a file containing newline-separated scrypted auth keys")
	flag.StringVar(&Config.remoteAuthFile, "remoteauthfile", "",
		"path to a file containing newline-separated scrypted auth keys for remote uploads")
	flag.StringVar(&Config.contentSecurityPolicy, "contentsecuritypolicy",
		"default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; referrer origin;",
		"value of default Content-Security-Policy header")
	flag.StringVar(&Config.fileContentSecurityPolicy, "filecontentsecuritypolicy",
		"default-src 'none'; img-src 'self'; object-src 'self'; media-src 'self'; sandbox; referrer origin;",
		"value of Content-Security-Policy header for file access")
	flag.StringVar(&Config.xFrameOptions, "xframeoptions", "SAMEORIGIN",
		"value of X-Frame-Options header")
	flag.Parse()

	mux := setup()

	if Config.fastcgi {
		listener, err := net.Listen("tcp", Config.bind)
		if err != nil {
			log.Fatal("Could not bind: ", err)
		}

		log.Printf("Serving over fastcgi, bound on %s, using siteurl %s", Config.bind, Config.siteURL)
		fcgi.Serve(listener, mux)
	} else if Config.certFile != "" {
		log.Printf("Serving over https, bound on %s, using siteurl %s", Config.bind, Config.siteURL)
		err := graceful.ListenAndServeTLS(Config.bind, Config.certFile, Config.keyFile, mux)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("Serving over http, bound on %s, using siteurl %s", Config.bind, Config.siteURL)
		err := graceful.ListenAndServe(Config.bind, mux)
		if err != nil {
			log.Fatal(err)
		}
	}
}
