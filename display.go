package main

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/andreimarcu/linx-server/expiry"
	"github.com/dustin/go-humanize"
	"github.com/flosch/pongo2"
	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
	"github.com/zenazn/goji/web"
)

const maxDisplayFileSizeBytes = 1024 * 512

var cliUserAgentRe = regexp.MustCompile("(?i)(lib)?curl|wget")

func fileDisplayHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	if !Config.noDirectAgents && cliUserAgentRe.MatchString(r.Header.Get("User-Agent")) && !strings.EqualFold("application/json", r.Header.Get("Accept")) {
		fileServeHandler(c, w, r)
		return
	}

	fileName := c.URLParams["name"]

	_, err := checkFile(fileName)
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
	if metadata.Expiry != expiry.NeverExpire {
		expiryHuman = humanize.RelTime(time.Now(), metadata.Expiry, "", "")
	}
	sizeHuman := humanize.Bytes(uint64(metadata.Size))
	extra := make(map[string]string)
	lines := []string{}

	extension := strings.TrimPrefix(filepath.Ext(fileName), ".")

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		js, _ := json.Marshal(map[string]string{
			"filename":   fileName,
			"direct_url": getSiteURL(r) + Config.selifPath + fileName,
			"expiry":     strconv.FormatInt(metadata.Expiry.Unix(), 10),
			"size":       strconv.FormatInt(metadata.Size, 10),
			"mimetype":   metadata.Mimetype,
			"sha256sum":  metadata.Sha256sum,
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

	} else if extension == "story" {
		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := fileBackend.Get(fileName)
			if err == nil {
				extra["contents"] = string(bytes)
				lines = strings.Split(extra["contents"], "\n")
				tpl = Templates["display/story.html"]
			}
		}

	} else if extension == "md" {
		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := fileBackend.Get(fileName)
			if err == nil {
				unsafe := blackfriday.MarkdownCommon(bytes)
				html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

				extra["contents"] = string(html)
				tpl = Templates["display/md.html"]
			}
		}

	} else if strings.HasPrefix(metadata.Mimetype, "text/") || supportedBinExtension(extension) {
		if metadata.Size < maxDisplayFileSizeBytes {
			bytes, err := fileBackend.Get(fileName)
			if err == nil {
				extra["extension"] = extension
				extra["lang_hl"], extra["lang_ace"] = extensionToHlAndAceLangs(extension)
				extra["contents"] = string(bytes)
				tpl = Templates["display/bin.html"]
			}
		}
	}

	// Catch other files
	if tpl == nil {
		tpl = Templates["display/file.html"]
	}

	err = renderTemplate(tpl, pongo2.Context{
		"mime":     metadata.Mimetype,
		"filename": fileName,
		"size":     sizeHuman,
		"expiry":   expiryHuman,
		"extra":    extra,
		"lines":    lines,
		"files":    metadata.ArchiveFiles,
	}, r, w)

	if err != nil {
		oopsHandler(c, w, r, RespHTML, "")
	}
}
