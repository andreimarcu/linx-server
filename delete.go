package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/zenazn/goji/web"
)

func deleteHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	requestKey := r.Header.Get("Linx-Delete-Key")

	filename := c.URLParams["name"]

	// Ensure requested file actually exists
	if _, readErr := fileBackend.Exists(filename); os.IsNotExist(readErr) {
		notFoundHandler(c, w, r) // 404 - file doesn't exist
		return
	}

	// Ensure delete key is correct
	metadata, err := metadataRead(filename)
	if err != nil {
		unauthorizedHandler(c, w, r) // 401 - no metadata available
		return
	}

	if metadata.DeleteKey == requestKey {
		fileDelErr := fileBackend.Delete(filename)
		metaDelErr := metaStorageBackend.Delete(filename)

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
