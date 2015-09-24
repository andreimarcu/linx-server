package main

import (
	"net/http"
	"os"
	"path"

	"github.com/zenazn/goji/web"
)

func fileServeHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	filename := c.URLParams["name"]
	absPath := path.Join(Config.filesDir, filename)
	_, err := os.Stat(absPath)

	if os.IsNotExist(err) {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	// plug file expiry checking here

	http.ServeFile(w, r, absPath)
}
