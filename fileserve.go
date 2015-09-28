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
	_, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		notFoundHandler(c, w, r)
		return
	}

	expired, expErr := isFileExpired(fileName)

	if expErr != nil {
		// Error reading metadata, pretend it's expired
		notFoundHandler(c, w, r)
		// TODO log error internally
		return
	} else if expired {
		notFoundHandler(c, w, r)
		// TODO delete the file
	}

	http.ServeFile(w, r, filePath)
}
