package main

import (
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/rakyll/magicmime"
	"github.com/zenazn/goji/web"
)

func fileDisplayHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fileName := c.URLParams["name"]
	filePath := path.Join(Config.filesDir, fileName)
	fileInfo, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		http.Error(w, http.StatusText(404), 404)
		return
	}

	if err := magicmime.Open(magicmime.MAGIC_MIME_TYPE |
		magicmime.MAGIC_SYMLINK |
		magicmime.MAGIC_ERROR); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	defer magicmime.Close()

	mimetype, err := magicmime.TypeByFile(filePath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var tpl *pongo2.Template

	if strings.HasPrefix(mimetype, "image/") {
		tpl = pongo2.Must(pongo2.FromCache("templates/display/image.html"))
	} else if  strings.HasPrefix(mimetype, "video/") {
		tpl = pongo2.Must(pongo2.FromCache("templates/display/video.html"))
	} else {
		tpl = pongo2.Must(pongo2.FromCache("templates/display/file.html"))
	}

	err = tpl.ExecuteWriter(pongo2.Context{
		"mime":     mimetype,
		"sitename": Config.siteName,
		"filename": fileName,
		"size":     fileInfo.Size(),
	}, w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
