package main

import (
	"fmt"
	"net/http"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/zenazn/goji/web"
)

func deleteHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	requestKey := r.Header.Get("Linx-Delete-Key")

	filename := c.URLParams["name"]

	// Ensure that file exists and delete key is correct
	metadata, err := storageBackend.Head(filename)
	if err == backends.NotFoundErr {
		notFoundHandler(c, w, r) // 404 - file doesn't exist
		return
	} else if err != nil {
		unauthorizedHandler(c, w, r) // 401 - no metadata available
		return
	}

	if metadata.DeleteKey == requestKey {
		err := storageBackend.Delete(filename)
		if err != nil {
			oopsHandler(c, w, r, RespPLAIN, "Could not delete")
			return
		}

		_, _ = fmt.Fprintf(w, "DELETED")
		return

	} else {
		unauthorizedHandler(c, w, r) // 401 - wrong delete key
		return
	}
}
