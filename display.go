package main

import (
	"net/http"
	"os"
	"path"
	"strings"

	"bitbucket.org/taruti/mimemagic"
	"github.com/flosch/pongo2"
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

	file, _ := os.Open(filePath)
	header := make([]byte, 512)
	file.Read(header)
	file.Close()

	mimetype := mimemagic.Match("", header)

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
