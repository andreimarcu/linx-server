package main

import (
	"net/http"

	"github.com/flosch/pongo2"
	"github.com/zenazn/goji/web"
)

func indexHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	indexTpl := pongo2.Must(pongo2.FromCache("templates/index.html"))

	err := indexTpl.ExecuteWriter(pongo2.Context{}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
