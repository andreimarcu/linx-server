package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"code.google.com/p/go-uuid/uuid"
	"github.com/zenazn/goji/web"
)

type UploadRequest struct {
	src            io.Reader
	filename       string
	expiry         int
	randomBarename bool
}

type Upload struct {
	Filename string
	Size     int64
	Expiry   int
}

func uploadPostHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	upReq := UploadRequest{}

	if r.Header.Get("Content-Type") == "application/octet-stream" {
		defer r.Body.Close()
		upReq.src = r.Body
		upReq.filename = r.URL.Query().Get("qqfile")

	} else {
		file, headers, err := r.FormFile("file")
		if err != nil {
			oopsHandler(c, w, r)
			return
		}
		defer file.Close()

		upReq.src = file
		upReq.filename = headers.Filename
	}

	upload, err := processUpload(upReq)
	if err != nil {
		oopsHandler(c, w, r)
		return
	}

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		js, _ := json.Marshal(map[string]string{
			"filename": upload.Filename,
			"url":      Config.siteURL + upload.Filename,
		})

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(js)

	} else {
		http.Redirect(w, r, "/"+upload.Filename, 301)
	}

}

func uploadPutHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	upReq := UploadRequest{}

	defer r.Body.Close()
	upReq.filename = c.URLParams["name"]
	upReq.src = r.Body

	upload, err := processUpload(upReq)
	if err != nil {
		oopsHandler(c, w, r)
		return
	}

	fmt.Fprintf(w, Config.siteURL+upload.Filename)
}

func processUpload(upReq UploadRequest) (upload Upload, err error) {
	barename, extension := barePlusExt(upReq.filename)

	if upReq.randomBarename || len(barename) == 0 {
		barename = generateBarename()
	}

	if len(extension) == 0 {
		extension = "ext"
	}

	upload.Filename = strings.Join([]string{barename, extension}, ".")

	_, err = os.Stat(path.Join(Config.filesDir, upload.Filename))

	fileexists := err == nil
	for fileexists {
		counter, err := strconv.Atoi(string(barename[len(barename)-1]))
		if err != nil {
			barename = barename + "1"
		} else {
			barename = barename[:len(barename)-1] + strconv.Itoa(counter+1)
		}
		upload.Filename = strings.Join([]string{barename, extension}, ".")

		_, err = os.Stat(path.Join(Config.filesDir, upload.Filename))
		fileexists = err == nil
	}

	dst, err := os.Create(path.Join(Config.filesDir, upload.Filename))
	if err != nil {
		return
	}
	defer dst.Close()

	bytes, err := io.Copy(dst, upReq.src)
	if err != nil {
		return
	} else if bytes == 0 {
		return
	}

	upload.Size = bytes
	return
}

func generateBarename() string {
	return uuid.New()[:8]
}

var barePlusRe = regexp.MustCompile(`[^A-Za-z0-9\-]`)

func barePlusExt(filename string) (barename, extension string) {

	filename = strings.TrimSpace(filename)
	filename = strings.ToLower(filename)

	extension = path.Ext(filename)
	barename = filename[:len(filename)-len(extension)]

	extension = barePlusRe.ReplaceAllString(extension, "")
	barename = barePlusRe.ReplaceAllString(barename, "")

	return
}
