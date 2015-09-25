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

var imageTpl = pongo2.Must(pongo2.FromCache("templates/display/image.html"))
var videoTpl = pongo2.Must(pongo2.FromCache("templates/display/video.html"))
var fileTpl = pongo2.Must(pongo2.FromCache("templates/display/file.html"))

func fileDisplayHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fileName := c.URLParams["name"]
	filePath := path.Join(Config.filesDir, fileName)
	fileInfo, err := os.Stat(filePath)

	if os.IsNotExist(err) {
		notFoundHandler(c, w, r)
		return
	}

	if err := magicmime.Open(magicmime.MAGIC_MIME_TYPE |
		magicmime.MAGIC_SYMLINK |
		magicmime.MAGIC_ERROR); err != nil {
		oopsHandler(c, w, r)
	}
	defer magicmime.Close()

	mimetype, err := magicmime.TypeByFile(filePath)
	if err != nil {
		oopsHandler(c, w, r)
	}

	var tpl *pongo2.Template

	if strings.HasPrefix(mimetype, "image/") {
		tpl = imageTpl
	} else if strings.HasPrefix(mimetype, "video/") {
		tpl = videoTpl
	} else {
		tpl = fileTpl
	}

	err = tpl.ExecuteWriter(pongo2.Context{
		"mime":     mimetype,
		"filename": fileName,
		"size":     fileInfo.Size(),
	}, w)

	if err != nil {
		oopsHandler(c, w, r)
	}
}
