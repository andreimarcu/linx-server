package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/zenazn/goji/web"
)

func deleteHandler(c web.C, w http.ResponseWriter, r *http.Request) {
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

	if deleteAllowed(r, metadata) {
		err := storageBackend.Delete(filename)
		if err != nil {
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

func deleteAllowed(r *http.Request, metadata backends.Metadata) bool {
	requestKey := r.Header.Get("Linx-Delete-Key")
	if metadata.DeleteKey == requestKey {
		return true
	}

	if Config.masterDeleteIp != "" {
		whitelist := strings.Split(Config.masterDeleteIp, ",")
		return remoteAddrInWhitelist(r, whitelist)
	}

	return false
}

func remoteAddrInWhitelist(r *http.Request, whitelist []string) bool {
	remoteAddr := strings.SplitN(r.RemoteAddr, ":", 2)[0]
	for _, ip := range whitelist {
		if ip == remoteAddr {
			return true
		}
	}
	return false
}
