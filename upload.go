package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
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

	file, headers, err := r.FormFile("file")
	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}
	defer file.Close()

	upReq.src = file
	upReq.filename = headers.Filename

	upload, err := processUpload(upReq)
	if err != nil {
		fmt.Fprintf(w, "Failed to upload: %v", err)
		return
	}

	fmt.Fprintf(w, "File %s uploaded successfully.", upload.Filename)
}

func uploadPutHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	upReq := UploadRequest{}

	defer r.Body.Close()
	upReq.src = r.Body

	upload, err := processUpload(upReq)
	if err != nil {
		fmt.Fprintf(w, "Failed to upload")
		return
	}

	fmt.Fprintf(w, "File %s uploaded successfully.", upload.Filename)
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

	dst, err := os.Create(path.Join("files/", upload.Filename))
	if err != nil {
		return
	}
	defer dst.Close()

	bytes, err := io.Copy(dst, upReq.src)
	if err != nil {
		return
	} else if bytes == 0 {
		err = errors.New("Empty file")
		return
	}

	upload.Size = bytes

	return
}

func generateBarename() string {
	return uuid.New()[:8]
}

func barePlusExt(filename string) (barename, extension string) {
	re := regexp.MustCompile(`[^A-Za-z0-9\-]`)

	filename = strings.TrimSpace(filename)
	filename = strings.ToLower(filename)

	extension = path.Ext(filename)
	barename = filename[:len(filename)-len(extension)]

	extension = re.ReplaceAllString(extension, "")
	barename = re.ReplaceAllString(barename, "")

	return
}
