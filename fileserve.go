package main

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/andreimarcu/linx-server/expiry"
	"github.com/zenazn/goji/web"
)

func fileServeHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fileName := c.URLParams["name"]

	metadata, err := checkFile(fileName)
	if err == NotFoundErr {
		notFoundHandler(c, w, r)
		return
	} else if err == backends.BadMetadata {
		oopsHandler(c, w, r, RespAUTO, "Corrupt metadata.")
		return
	} else if err != nil {
		oopsHandler(c, w, r, RespAUTO, err.Error())
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

	w.Header().Set("Etag", metadata.Sha256sum)
	w.Header().Set("Cache-Control", "max-age=0")

	fileBackend.ServeFile(fileName, w, r)
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

		w.Header().Set("Etag", timeStartedStr)
		w.Header().Set("Cache-Control", "max-age=86400")
		http.ServeContent(w, r, filePath, timeStarted, file)
		return
	}
}

func checkFile(filename string) (metadata backends.Metadata, err error) {
	_, err = fileBackend.Exists(filename)
	if err != nil {
		err = NotFoundErr
		return
	}

	metadata, err = metadataRead(filename)
	if err != nil {
		return
	}

	if expiry.IsTsExpired(metadata.Expiry) {
		fileBackend.Delete(filename)
		metaStorageBackend.Delete(filename)
		err = NotFoundErr
		return
	}

	return
}
