package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/taruti/mimemagic"
	"github.com/dustin/go-humanize"
	"github.com/flosch/pongo2"
	"github.com/zenazn/goji/web"
)

const maxDisplayFileSizeBytes = 1024 * 512

func fileDisplayHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fileName := c.URLParams["name"]
	filePath := path.Join(Config.filesDir, fileName)
	fileInfo, err := os.Stat(filePath)

	if !fileExistsAndNotExpired(fileName) {
		notFoundHandler(c, w, r)
		return
	}

	expiry, _ := metadataGetExpiry(fileName)
	var expiryHuman string
	if expiry != neverExpire {
		expiryHuman = humanize.RelTime(time.Now(), expiry, "", "")
	}
	sizeHuman := humanize.Bytes(uint64(fileInfo.Size()))
	extra := make(map[string]string)

	file, _ := os.Open(filePath)
	header := make([]byte, 512)
	file.Read(header)
	file.Close()

	mimetype := mimemagic.Match("", header)
	extension := strings.TrimPrefix(filepath.Ext(fileName), ".")

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		js, _ := json.Marshal(map[string]string{
			"filename": fileName,
			"mimetype": mimetype,
			"expiry":   strconv.FormatInt(expiry.Unix(), 10),
			"size":     strconv.FormatInt(fileInfo.Size(), 10),
		})
		w.Write(js)
		return
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
	} else if supportedBinExtension(extension) {
		if fileInfo.Size() < maxDisplayFileSizeBytes {
			bytes, err := ioutil.ReadFile(filePath)
			if err != nil {
				tpl = Templates["display/file.html"]
			} else {
				extra["extension"] = extension
				extra["lang_hl"], extra["lang_ace"] = extensionToHlAndAceLangs(extension)
				extra["contents"] = string(bytes)
				tpl = Templates["display/bin.html"]
			}
		} else {
			tpl = Templates["display/file.html"]
		}
	} else {
		tpl = Templates["display/file.html"]
	}

	err = tpl.ExecuteWriter(pongo2.Context{
		"mime":     mimetype,
		"filename": fileName,
		"size":     sizeHuman,
		"expiry":   expiryHuman,
		"extra":    extra,
	}, w)

	if err != nil {
		oopsHandler(c, w, r, RespHTML, "")
	}
}
