package main

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/zenazn/goji/web"
)

func deleteHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	requestKey := r.Header.Get("X-Delete-Key")

	filename := c.URLParams["name"]
	filePath := path.Join(Config.filesDir, filename)
	metaPath := path.Join(Config.metaDir, filename)

	// Ensure requested file actually exists
	if _, readErr := os.Stat(filePath); os.IsNotExist(readErr) {
		notFoundHandler(c, w, r) // 404 - file doesn't exist
		return
	}

	// Ensure delete key is correct
	deleteKey, err := metadataGetDeleteKey(filename)

	if err != nil {
		unauthorizedHandler(c, w, r) // 401 - no metadata available
		return
	}

	if deleteKey == requestKey {
		fileDelErr := os.Remove(filePath)
		metaDelErr := os.Remove(metaPath)

		if (fileDelErr != nil) || (metaDelErr != nil) {
			oopsHandler(c, w, r, RespPLAIN, "Could not delete")
			return
		}

		fmt.Fprintf(w, "DELETED")
		return

	} else {
		unauthorizedHandler(c, w, r) // 401 - wrong delete key
		return
	}
}
