package main

import (
	"flag"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/andreimarcu/linx-server/backends"
	"github.com/andreimarcu/linx-server/backends/localfs"
	"github.com/andreimarcu/linx-server/backends/metajson"
	"github.com/flosch/pongo2"
	"github.com/vharitonsky/iniflags"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"
)

type headerList []string

func (h *headerList) String() string {
	return strings.Join(*h, ",")
}

func (h *headerList) Set(value string) error {
	*h = append(*h, value)
	return nil
}

var Config struct {
	bind                      string
	filesDir                  string
	metaDir                   string
	siteName                  string
	siteURL                   string
	sitePath                  string
	certFile                  string
	keyFile                   string
	contentSecurityPolicy     string
	fileContentSecurityPolicy string
	xFrameOptions             string
	maxSize                   int64
	maxExpiry                 uint64
	realIp                    bool
	noLogs                    bool
	allowHotlink              bool
	fastcgi                   bool
	remoteUploads             bool
	authFile                  string
	remoteAuthFile            string
	addHeaders                headerList
	googleShorterAPIKey       string
	noDirectAgents            bool
}

var Templates = make(map[string]*pongo2.Template)
var TemplateSet *pongo2.TemplateSet
var staticBox *rice.Box
var timeStarted time.Time
var timeStartedStr string
var remoteAuthKeys []string
var metaStorageBackend backends.MetaStorageBackend
var metaBackend backends.MetaBackend
var fileBackend backends.StorageBackend

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
	mux.Use(AddHeaders(Config.addHeaders))

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

	if Config.siteURL != "" {
		// ensure siteURL ends wth '/'
		if lastChar := Config.siteURL[len(Config.siteURL)-1:]; lastChar != "/" {
			Config.siteURL = Config.siteURL + "/"
		}

		parsedUrl, err := url.Parse(Config.siteURL)
		if err != nil {
			log.Fatal("Could not parse siteurl:", err)
		}

		Config.sitePath = parsedUrl.Path
	} else {
		Config.sitePath = "/"
	}

	metaStorageBackend = localfs.NewLocalfsBackend(Config.metaDir)
	metaBackend = metajson.NewMetaJSONBackend(metaStorageBackend)
	fileBackend = localfs.NewLocalfsBackend(Config.filesDir)

	// Template setup
	p2l, err := NewPongo2TemplatesLoader()
	if err != nil {
		log.Fatal("Error: could not load templates", err)
	}
	TemplateSet := pongo2.NewSet("templates", p2l)
	err = populateTemplatesMap(TemplateSet, Templates)
	if err != nil {
		log.Fatal("Error: could not load templates", err)
	}

	staticBox = rice.MustFindBox("static")
	timeStarted = time.Now()
	timeStartedStr = strconv.FormatInt(timeStarted.Unix(), 10)

	// Routing setup
	nameRe := regexp.MustCompile("^" + Config.sitePath + `(?P<name>[a-z0-9-\.]+)$`)
	selifRe := regexp.MustCompile("^" + Config.sitePath + `selif/(?P<name>[a-z0-9-\.]+)$`)
	selifIndexRe := regexp.MustCompile("^" + Config.sitePath + `selif/$`)
	torrentRe := regexp.MustCompile("^" + Config.sitePath + `(?P<name>[a-z0-9-\.]+)/torrent$`)
	shortRe := regexp.MustCompile("^" + Config.sitePath + `(?P<name>[a-z0-9-\.]+)/short$`)

	if Config.authFile == "" {
		mux.Get(Config.sitePath, indexHandler)
		mux.Get(Config.sitePath+"paste/", pasteHandler)
	} else {
		mux.Get(Config.sitePath, http.RedirectHandler(Config.sitePath+"API", 303))
		mux.Get(Config.sitePath+"paste/", http.RedirectHandler(Config.sitePath+"API/", 303))
	}
	mux.Get(Config.sitePath+"paste", http.RedirectHandler(Config.sitePath+"paste/", 301))

	mux.Get(Config.sitePath+"API/", apiDocHandler)
	mux.Get(Config.sitePath+"API", http.RedirectHandler(Config.sitePath+"API/", 301))

	if Config.remoteUploads {
		mux.Get(Config.sitePath+"upload", uploadRemote)
		mux.Get(Config.sitePath+"upload/", uploadRemote)

		if Config.remoteAuthFile != "" {
			remoteAuthKeys = readAuthKeys(Config.remoteAuthFile)
		}
	}

	mux.Post(Config.sitePath+"upload", uploadPostHandler)
	mux.Post(Config.sitePath+"upload/", uploadPostHandler)
	mux.Put(Config.sitePath+"upload", uploadPutHandler)
	mux.Put(Config.sitePath+"upload/", uploadPutHandler)
	mux.Put(Config.sitePath+"upload/:name", uploadPutHandler)

	mux.Delete(Config.sitePath+":name", deleteHandler)

	mux.Get(Config.sitePath+"static/*", staticHandler)
	mux.Get(Config.sitePath+"favicon.ico", staticHandler)
	mux.Get(Config.sitePath+"robots.txt", staticHandler)
	mux.Get(nameRe, fileDisplayHandler)
	mux.Get(selifRe, fileServeHandler)
	mux.Get(selifIndexRe, unauthorizedHandler)
	mux.Get(torrentRe, fileTorrentHandler)

	if Config.googleShorterAPIKey != "" {
		mux.Get(shortRe, shortURLHandler)
	}

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
	flag.StringVar(&Config.siteName, "sitename", "",
		"name of the site")
	flag.StringVar(&Config.siteURL, "siteurl", "",
		"site base url (including trailing slash)")
	flag.Int64Var(&Config.maxSize, "maxsize", 4*1024*1024*1024,
		"maximum upload file size in bytes (default 4GB)")
	flag.Uint64Var(&Config.maxExpiry, "maxexpiry", 0,
		"maximum expiration time in seconds (default is 0, which is no expiry)")
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
		"default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'self'; referrer origin;",
		"value of default Content-Security-Policy header")
	flag.StringVar(&Config.fileContentSecurityPolicy, "filecontentsecuritypolicy",
		"default-src 'none'; img-src 'self'; object-src 'self'; media-src 'self'; style-src 'self' 'unsafe-inline'; frame-ancestors 'self'; referrer origin;",
		"value of Content-Security-Policy header for file access")
	flag.StringVar(&Config.xFrameOptions, "xframeoptions", "SAMEORIGIN",
		"value of X-Frame-Options header")
	flag.Var(&Config.addHeaders, "addheader",
		"Add an arbitrary header to the response. This option can be used multiple times.")
	flag.StringVar(&Config.googleShorterAPIKey, "googleapikey", "",
		"API Key for Google's URL Shortener.")
	flag.BoolVar(&Config.noDirectAgents, "nodirectagents", false,
		"disable serving files directly for wget/curl user agents")

	iniflags.Parse()

	mux := setup()

	if Config.fastcgi {
		listener, err := net.Listen("tcp", Config.bind)
		if err != nil {
			log.Fatal("Could not bind: ", err)
		}

		log.Printf("Serving over fastcgi, bound on %s", Config.bind)
		fcgi.Serve(listener, mux)
	} else if Config.certFile != "" {
		log.Printf("Serving over https, bound on %s", Config.bind)
		err := graceful.ListenAndServeTLS(Config.bind, Config.certFile, Config.keyFile, mux)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		log.Printf("Serving over http, bound on %s", Config.bind)
		err := graceful.ListenAndServe(Config.bind, mux)
		if err != nil {
			log.Fatal(err)
		}
	}
}
