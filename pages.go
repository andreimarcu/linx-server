package main

import (
	"net/http"

	"github.com/flosch/pongo2"
	"github.com/zenazn/goji/web"
)

func indexHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	err := Templates["index.html"].ExecuteWriter(pongo2.Context{}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func notFoundHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	err := Templates["404.html"].ExecuteWriter(pongo2.Context{}, w)
	if err != nil {
		oopsHandler(c, w, r)
	}
}

func oopsHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	err := Templates["oops.html"].ExecuteWriter(pongo2.Context{}, w)
	if err != nil {
		oopsHandler(c, w, r)
	}
}
