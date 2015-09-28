package main

import (
	"net/http"
	"os"
	"path"

	"github.com/zenazn/goji/web"
)

func fileServeHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fileName := c.URLParams["name"]
	filePath := path.Join(Config.filesDir, fileName)

	if isFileExpired(fileName) {
		notFoundHandler(c, w, r)
		return
	}

	http.ServeFile(w, r, filePath)
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
