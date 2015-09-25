package main

import (
	"net/http"

	"github.com/flosch/pongo2"
	"github.com/zenazn/goji/web"
)

var indexTpl = pongo2.Must(pongo2.FromCache("templates/index.html"))
var notFoundTpl = pongo2.Must(pongo2.FromCache("templates/404.html"))
var oopsTpl = pongo2.Must(pongo2.FromCache("templates/oops.html"))

func indexHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	err := indexTpl.ExecuteWriter(pongo2.Context{}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func notFoundHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	err := notFoundTpl.ExecuteWriter(pongo2.Context{}, w)
	if err != nil {
		oopsHandler(c, w, r)
	}
}

func oopsHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	err := oopsTpl.ExecuteWriter(pongo2.Context{}, w)
	if err != nil {
		oopsHandler(c, w, r)
	}
}
