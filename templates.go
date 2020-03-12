package main

import (
	"bytes"
	"io"
	"net/http"
	"path"
	"path/filepath"
	"strings"

	rice "github.com/GeertJohan/go.rice"
	"github.com/flosch/pongo2"
)

type Pongo2Loader struct {
	box *rice.Box
}

func NewPongo2TemplatesLoader() (*Pongo2Loader, error) {
	fs := &Pongo2Loader{}

	p2l, err := rice.FindBox("templates")
	if err != nil {
		return nil, err
	}

	fs.box = p2l
	return fs, nil
}

func (fs *Pongo2Loader) Get(path string) (io.Reader, error) {
	myBytes, err := fs.box.Bytes(path)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(myBytes), nil
}

func (fs *Pongo2Loader) Abs(base, name string) string {
	me := path.Join(filepath.Dir(base), name)
	return me
}

func populateTemplatesMap(tSet *pongo2.TemplateSet, tMap map[string]*pongo2.Template) error {
	templates := []string{
		"index.html",
		"paste.html",
		"API.html",
		"400.html",
		"401.html",
		"404.html",
		"oops.html",
		"access.html",
		"custom_page.html",

		"display/audio.html",
		"display/image.html",
		"display/video.html",
		"display/pdf.html",
		"display/bin.html",
		"display/story.html",
		"display/md.html",
		"display/file.html",
	}

	for _, tName := range templates {
		tpl, err := tSet.FromFile(tName)
		if err != nil {
			return err
		}

		tMap[tName] = tpl
	}

	return nil
}

func renderTemplate(tpl *pongo2.Template, context pongo2.Context, r *http.Request, writer io.Writer) error {
	if Config.siteName == "" {
		parts := strings.Split(r.Host, ":")
		context["sitename"] = parts[0]
	} else {
		context["sitename"] = Config.siteName
	}

	context["sitepath"] = Config.sitePath
	context["selifpath"] = Config.selifPath
	context["custom_pages_names"] = customPagesNames

	var a string
	if Config.authFile == "" {
		a = "none"
	} else if Config.basicAuth {
		a = "basic"
	} else {
		a = "header"
	}
	context["auth"] = a

	return tpl.ExecuteWriter(context, writer)
}
