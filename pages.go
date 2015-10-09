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
	err := Templates["index.html"].ExecuteWriter(pongo2.Context{
		"maxsize": Config.maxSize,
	}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func pasteHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	err := Templates["paste.html"].ExecuteWriter(pongo2.Context{}, w)
	if err != nil {
		oopsHandler(c, w, r, RespHTML, "")
	}
}

func apiDocHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	err := Templates["api.html"].ExecuteWriter(pongo2.Context{}, w)
	if err != nil {
		oopsHandler(c, w, r, RespHTML, "")
	}
}

func notFoundHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	err := Templates["404.html"].ExecuteWriter(pongo2.Context{}, w)
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
		Templates["oops.html"].ExecuteWriter(pongo2.Context{"msg": msg}, w)
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

func unauthorizedHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(401)
	err := Templates["401.html"].ExecuteWriter(pongo2.Context{}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
