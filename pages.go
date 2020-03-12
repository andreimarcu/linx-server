package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/zenazn/goji/web"
)

type RespType int

const (
	RespPLAIN RespType = iota
	RespJSON
	RespHTML
	RespAUTO
)

func indexHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	err := renderTemplate(Templates["index.html"], pongo2.Context{
		"maxsize":     Config.maxSize,
		"expirylist":  listExpirationTimes(),
		"forcerandom": Config.forceRandomFilename,
	}, r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func pasteHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	err := renderTemplate(Templates["paste.html"], pongo2.Context{
		"expirylist":  listExpirationTimes(),
		"forcerandom": Config.forceRandomFilename,
	}, r, w)
	if err != nil {
		oopsHandler(c, w, r, RespHTML, "")
	}
}

func apiDocHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	err := renderTemplate(Templates["API.html"], pongo2.Context{
		"siteurl":     getSiteURL(r),
		"forcerandom": Config.forceRandomFilename,
	}, r, w)
	if err != nil {
		oopsHandler(c, w, r, RespHTML, "")
	}
}

func makeCustomPageHandler(fileName string) func(c web.C, w http.ResponseWriter, r *http.Request) {
	return func(c web.C, w http.ResponseWriter, r *http.Request) {
		err := renderTemplate(Templates["custom_page.html"], pongo2.Context{
			"siteurl":     getSiteURL(r),
			"forcerandom": Config.forceRandomFilename,
			"contents":    customPages[fileName],
			"filename":    fileName,
			"pagename":    customPagesNames[fileName],
		}, r, w)
		if err != nil {
			oopsHandler(c, w, r, RespHTML, "")
		}
	}
}

func notFoundHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	err := renderTemplate(Templates["404.html"], pongo2.Context{}, r, w)
	if err != nil {
		oopsHandler(c, w, r, RespHTML, "")
	}
}

func oopsHandler(c web.C, w http.ResponseWriter, r *http.Request, rt RespType, msg string) {
	if msg == "" {
		msg = "Oops! Something went wrong..."
	}

	if rt == RespHTML {
		w.WriteHeader(500)
		renderTemplate(Templates["oops.html"], pongo2.Context{"msg": msg}, r, w)
		return
	} else if rt == RespPLAIN {
		w.WriteHeader(500)
		fmt.Fprintf(w, "%s", msg)
		return
	} else if rt == RespJSON {
		js, _ := json.Marshal(map[string]string{
			"error": msg,
		})

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(500)
		w.Write(js)
		return
	} else if rt == RespAUTO {
		if strings.EqualFold("application/json", r.Header.Get("Accept")) {
			oopsHandler(c, w, r, RespJSON, msg)
		} else {
			oopsHandler(c, w, r, RespHTML, msg)
		}
	}
}

func badRequestHandler(c web.C, w http.ResponseWriter, r *http.Request, rt RespType, msg string) {
	if rt == RespHTML {
		w.WriteHeader(http.StatusBadRequest)
		err := renderTemplate(Templates["400.html"], pongo2.Context{"msg": msg}, r, w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	} else if rt == RespPLAIN {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "%s", msg)
		return
	} else if rt == RespJSON {
		js, _ := json.Marshal(map[string]string{
			"error": msg,
		})

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(js)
		return
	} else if rt == RespAUTO {
		if strings.EqualFold("application/json", r.Header.Get("Accept")) {
			badRequestHandler(c, w, r, RespJSON, msg)
		} else {
			badRequestHandler(c, w, r, RespHTML, msg)
		}
	}
}

func unauthorizedHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(401)
	err := renderTemplate(Templates["401.html"], pongo2.Context{}, r, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
