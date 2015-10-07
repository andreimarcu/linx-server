package main

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"bitbucket.org/taruti/mimemagic"
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
	files := []string{}

	file, _ := os.Open(filePath)
	defer file.Close()

	header := make([]byte, 512)
	file.Read(header)

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
	} else if mimetype == "application/x-tar" {
		f, _ := os.Open(filePath)
		defer f.Close()

		tReadr := tar.NewReader(f)
		for {
			header, err := tReadr.Next()
			if err == io.EOF || err != nil {
				break
			}

			if header.Typeflag == tar.TypeDir || header.Typeflag == tar.TypeReg {
				files = append(files, header.Name)
			}
		}
		sort.Strings(files)

	} else if mimetype == "application/x-gzip" {
		f, _ := os.Open(filePath)
		defer f.Close()

		gzf, err := gzip.NewReader(f)
		if err == nil {
			tReadr := tar.NewReader(gzf)
			for {
				header, err := tReadr.Next()
				if err == io.EOF || err != nil {
					break
				}

				if header.Typeflag == tar.TypeDir || header.Typeflag == tar.TypeReg {
					files = append(files, header.Name)
				}
			}
			sort.Strings(files)
		}
	} else if mimetype == "application/x-bzip" {
		f, _ := os.Open(filePath)
		defer f.Close()

		bzf := bzip2.NewReader(f)
		tReadr := tar.NewReader(bzf)
		for {
			header, err := tReadr.Next()
			if err == io.EOF || err != nil {
				break
			}

			if header.Typeflag == tar.TypeDir || header.Typeflag == tar.TypeReg {
				files = append(files, header.Name)
			}
		}
		sort.Strings(files)

	} else if mimetype == "application/zip" {
		f, _ := os.Open(filePath)
		defer f.Close()

		zf, err := zip.NewReader(f, fileInfo.Size())
		if err == nil {
			for _, f := range zf.File {
				files = append(files, f.Name)
			}
		}

	} else if supportedBinExtension(extension) {
		if fileInfo.Size() < maxDisplayFileSizeBytes {
			bytes, err := ioutil.ReadFile(filePath)
			if err == nil {
				extra["extension"] = extension
				extra["lang_hl"], extra["lang_ace"] = extensionToHlAndAceLangs(extension)
				extra["contents"] = string(bytes)
				tpl = Templates["display/bin.html"]
			}
		}
	} else if extension == "md" {
		if fileInfo.Size() < maxDisplayFileSizeBytes {
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
		"mime":     mimetype,
		"filename": fileName,
		"size":     sizeHuman,
		"expiry":   expiryHuman,
		"extra":    extra,
		"files":    files,
	}, w)

	if err != nil {
		oopsHandler(c, w, r, RespHTML, "")
	}
}
