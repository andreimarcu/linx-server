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

	if !fileExistsAndNotExpired(fileName) {
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
		tpl = Templates["display/image.html"]
	} else if strings.HasPrefix(mimetype, "video/") {
		tpl = Templates["display/video.html"]
	} else if strings.HasPrefix(mimetype, "audio/") {
		tpl = Templates["display/audio.html"]
	} else if mimetype == "application/pdf" {
		tpl = Templates["display/pdf.html"]
	} else {
		tpl = Templates["display/file.html"]
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
