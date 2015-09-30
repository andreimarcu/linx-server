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

func pasteHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	err := Templates["paste.html"].ExecuteWriter(pongo2.Context{}, w)
	if err != nil {
		oopsHandler(c, w, r)
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
	w.WriteHeader(500)
	err := Templates["oops.html"].ExecuteWriter(pongo2.Context{}, w)
	if err != nil {
		oopsHandler(c, w, r)
	}
}

func unauthorizedHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(401)
	err := Templates["401.html"].ExecuteWriter(pongo2.Context{}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
