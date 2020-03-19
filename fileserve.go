package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/andreimarcu/linx-server/expiry"
	"github.com/andreimarcu/linx-server/httputil"
	"github.com/zenazn/goji/web"
)

func fileServeHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fileName := c.URLParams["name"]

	metadata, err := checkFile(fileName)
	if err == backends.NotFoundErr {
		notFoundHandler(c, w, r)
		return
	} else if err != nil {
		oopsHandler(c, w, r, RespAUTO, "Corrupt metadata.")
		return
	}

	if src, err := checkAccessKey(r, &metadata); err != nil {
		// remove invalid cookie
		if src == accessKeySourceCookie {
			setAccessKeyCookies(w, getSiteURL(r), fileName, "", time.Unix(0, 0))
		}
		unauthorizedHandler(c, w, r)

		return
	}

	if !Config.allowHotlink {
		referer := r.Header.Get("Referer")
		u, _ := url.Parse(referer)
		p, _ := url.Parse(getSiteURL(r))
		if referer != "" && !sameOrigin(u, p) {
			http.Redirect(w, r, Config.sitePath+fileName, 303)
			return
		}
	}

	w.Header().Set("Content-Security-Policy", Config.fileContentSecurityPolicy)
	w.Header().Set("Referrer-Policy", Config.fileReferrerPolicy)

	w.Header().Set("Content-Type", metadata.Mimetype)
	w.Header().Set("Content-Length", strconv.FormatInt(metadata.Size, 10))
	w.Header().Set("Etag", fmt.Sprintf("\"%s\"", metadata.Sha256sum))
	w.Header().Set("Cache-Control", "public, no-cache")

	modtime := time.Unix(0, 0)
	if done := httputil.CheckPreconditions(w, r, modtime); done == true {
		return
	}

	if r.Method != "HEAD" {

		storageBackend.ServeFile(fileName, w, r)
		if err != nil {
			oopsHandler(c, w, r, RespAUTO, err.Error())
			return
		}
	}
}

func staticHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path[len(path)-1:] == "/" {
		notFoundHandler(c, w, r)
		return
	} else {
		if path == "/favicon.ico" {
			path = Config.sitePath + "/static/images/favicon.gif"
		}

		filePath := strings.TrimPrefix(path, Config.sitePath+"static/")
		file, err := staticBox.Open(filePath)
		if err != nil {
			notFoundHandler(c, w, r)
			return
		}

		w.Header().Set("Etag", fmt.Sprintf("\"%s\"", timeStartedStr))
		w.Header().Set("Cache-Control", "public, max-age=86400")
		http.ServeContent(w, r, filePath, timeStarted, file)
		return
	}
}

func checkFile(filename string) (metadata backends.Metadata, err error) {
	metadata, err = storageBackend.Head(filename)
	if err != nil {
		return
	}

	if expiry.IsTsExpired(metadata.Expiry) {
		storageBackend.Delete(filename)
		err = backends.NotFoundErr
		return
	}

	return
}
