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

	"github.com/dustin/go-humanize"
	"github.com/flosch/pongo2"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"github.com/zenazn/goji/web"
)

const maxDisplayFileSizeBytes = 1024 * 512

func fileDisplayHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fileName := c.URLParams["name"]
	filePath := path.Join(Config.filesDir, fileName)

	err := checkFile(fileName)
	if err == NotFoundErr {
		notFoundHandler(c, w, r)
		return
	}

	metadata, err := metadataRead(fileName)
	if err != nil {
		oopsHandler(c, w, r, RespAUTO, "Corrupt metadata.")
		return
	}
	var expiryHuman string
	if metadata.Expiry != neverExpire {
		expiryHuman = humanize.RelTime(time.Now(), metadata.Expiry, "", "")
	}
	sizeHuman := humanize.Bytes(uint64(metadata.Size))
	extra := make(map[string]string)
	lines := []string{}

	file, _ := os.Open(filePath)
	defer file.Close()

	header := make([]byte, 512)
	file.Read(header)

	extension := strings.TrimPrefix(filepath.Ext(fileName), ".")

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		js, _ := json.Marshal(map[string]string{
			"filename":  fileName,
			"expiry":    strconv.FormatInt(metadata.Expiry.Unix(), 10),
			"size":      strconv.FormatInt(metadata.Size, 10),
			"mimetype":  metadata.Mimetype,
			"sha256sum": metadata.Sha256sum,
		})
		w.Write(js)
		return
	}

	var tpl *pongo2.Template

	if strings.HasPrefix(metadata.Mimetype, "image/") {
		tpl = Templates["display/image.html"]

	} else if strings.HasPrefix(metadata.Mimetype, "video/") {
		tpl = Templates["display/video.html"]

	} else if strings.HasPrefix(metadata.Mimetype, "audio/") {
		tpl = Templates["display/audio.html"]

	} else if metadata.Mimetype == "application/pdf" {
		tpl = Templates["display/pdf.html"]

	} else if metadata.Mimetype == "text/plain" || supportedBinExtension(extension) {
		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := ioutil.ReadFile(filePath)
			if err == nil {
				extra["extension"] = extension
				extra["lang_hl"], extra["lang_ace"] = extensionToHlAndAceLangs(extension)
				extra["contents"] = string(bytes)
				tpl = Templates["display/bin.html"]
			}
		}
	} else if extension == "story" {
		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := ioutil.ReadFile(filePath)
			if err == nil {
				extra["contents"] = string(bytes)
				lines = strings.Split(extra["contents"], "\n")
				tpl = Templates["display/story.html"]
			}
		}
	} else if extension == "md" {
		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := ioutil.ReadFile(filePath)
			if err == nil {
				unsafe := blackfriday.MarkdownCommon(bytes)
				html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

				extra["contents"] = string(html)
				tpl = Templates["display/md.html"]
			}
		}
	}

	// Catch other files
	if tpl == nil {
		tpl = Templates["display/file.html"]
	}

	err = tpl.ExecuteWriter(pongo2.Context{
		"mime":     metadata.Mimetype,
		"filename": fileName,
		"size":     sizeHuman,
		"expiry":   expiryHuman,
		"extra":    extra,
		"lines":    lines,
		"files":    metadata.ArchiveFiles,
	}, w)

	if err != nil {
		oopsHandler(c, w, r, RespHTML, "")
	}
}
