package main

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/zenazn/goji/web"
)

func fileServeHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fileName := c.URLParams["name"]
	filePath := path.Join(Config.filesDir, fileName)

	if !fileExistsAndNotExpired(fileName) {
		notFoundHandler(c, w, r)
		return
	}

	if !Config.allowHotlink {
		referer := r.Header.Get("Referer")
		if referer != "" && !strings.HasPrefix(referer, Config.siteURL) {
			w.WriteHeader(403)
			return
		}
	}

	w.Header().Set("Content-Security-Policy", Config.fileContentSecurityPolicy)

	http.ServeFile(w, r, filePath)
}

func staticHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path[len(path)-1:] == "/" {
		notFoundHandler(c, w, r)
		return
	} else {
		if path == "/favicon.ico" {
			path = "/static/images/favicon.gif"
		}

		filePath := strings.TrimPrefix(path, "/static/")
		file, err := staticBox.Open(filePath)
		if err != nil {
			notFoundHandler(c, w, r)
			return
		}

		w.Header().Set("Etag", timeStartedStr)
		w.Header().Set("Cache-Control", "max-age=86400")
		http.ServeContent(w, r, filePath, timeStarted, file)
		return
	}
}

func fileExistsAndNotExpired(filename string) bool {
	filePath := path.Join(Config.filesDir, filename)

	_, err := os.Stat(filePath)
	if err != nil {
		return false
	}

	if isFileExpired(filename) {
		os.Remove(path.Join(Config.filesDir, filename))
		os.Remove(path.Join(Config.metaDir, filename))
		return false
	}

	return true
}
