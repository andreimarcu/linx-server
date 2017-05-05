package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/andreimarcu/linx-server/expiry"
	"github.com/dchest/uniuri"
	"github.com/zenazn/goji/web"
	"gopkg.in/h2non/filetype.v1"
)

var fileBlacklist = map[string]bool{
	"favicon.ico":     true,
	"index.htm":       true,
	"index.html":      true,
	"index.php":       true,
	"robots.txt":      true,
	"crossdomain.xml": true,
}

// Describes metadata directly from the user request
type UploadRequest struct {
	src            io.Reader
	filename       string
	expiry         time.Duration // Seconds until expiry, 0 = never
	randomBarename bool
	deletionKey    string // Empty string if not defined
}

// Metadata associated with a file as it would actually be stored
type Upload struct {
	Filename string // Final filename on disk
	Metadata backends.Metadata
}

func uploadPostHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	if !strictReferrerCheck(r, getSiteURL(r), []string{"Linx-Delete-Key", "Linx-Expiry", "Linx-Randomize", "X-Requested-With"}) {
		badRequestHandler(c, w, r)
		return
	}

	upReq := UploadRequest{}
	uploadHeaderProcess(r, &upReq)

	contentType := r.Header.Get("Content-Type")

	if strings.HasPrefix(contentType, "multipart/form-data") {
		file, headers, err := r.FormFile("file")
		if err != nil {
			oopsHandler(c, w, r, RespHTML, "Could not upload file.")
			return
		}
		defer file.Close()

		r.ParseForm()
		if r.Form.Get("randomize") == "true" {
			upReq.randomBarename = true
		}
		upReq.expiry = parseExpiry(r.Form.Get("expires"))
		upReq.src = file
		upReq.filename = headers.Filename
	} else {
		if r.FormValue("content") == "" {
			oopsHandler(c, w, r, RespHTML, "Empty file")
			return
		}
		extension := r.FormValue("extension")
		if extension == "" {
			extension = "txt"
		}

		upReq.src = strings.NewReader(r.FormValue("content"))
		upReq.expiry = parseExpiry(r.FormValue("expires"))
		upReq.filename = r.FormValue("filename") + "." + extension
	}

	upload, err := processUpload(upReq)

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		if err != nil {
			oopsHandler(c, w, r, RespJSON, "Could not upload file: "+err.Error())
			return
		}

		js := generateJSONresponse(upload, r)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(js)
	} else {
		if err != nil {
			oopsHandler(c, w, r, RespHTML, "Could not upload file: "+err.Error())
			return
		}

		http.Redirect(w, r, Config.sitePath+upload.Filename, 303)
	}

}

func uploadPutHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	upReq := UploadRequest{}
	uploadHeaderProcess(r, &upReq)

	defer r.Body.Close()
	upReq.filename = c.URLParams["name"]
	upReq.src = r.Body

	upload, err := processUpload(upReq)

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		if err != nil {
			oopsHandler(c, w, r, RespJSON, "Could not upload file: "+err.Error())
			return
		}

		js := generateJSONresponse(upload, r)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(js)
	} else {
		if err != nil {
			oopsHandler(c, w, r, RespPLAIN, "Could not upload file: "+err.Error())
			return
		}

		fmt.Fprintf(w, "%s\n", getSiteURL(r)+upload.Filename)
	}
}

func uploadRemote(c web.C, w http.ResponseWriter, r *http.Request) {
	if Config.remoteAuthFile != "" {
		result, err := checkAuth(remoteAuthKeys, r.FormValue("key"))
		if err != nil || !result {
			unauthorizedHandler(c, w, r)
			return
		}
	}

	if r.FormValue("url") == "" {
		http.Redirect(w, r, Config.sitePath, 303)
		return
	}

	upReq := UploadRequest{}
	grabUrl, _ := url.Parse(r.FormValue("url"))

	resp, err := http.Get(grabUrl.String())
	if err != nil {
		oopsHandler(c, w, r, RespAUTO, "Could not retrieve URL")
		return
	}

	upReq.filename = filepath.Base(grabUrl.Path)
	upReq.src = resp.Body
	upReq.deletionKey = r.FormValue("deletekey")
	upReq.randomBarename = r.FormValue("randomize") == "yes"
	upReq.expiry = parseExpiry(r.FormValue("expiry"))

	upload, err := processUpload(upReq)

	if strings.EqualFold("application/json", r.Header.Get("Accept")) {
		if err != nil {
			oopsHandler(c, w, r, RespJSON, "Could not upload file: "+err.Error())
			return
		}

		js := generateJSONresponse(upload, r)
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(js)
	} else {
		if err != nil {
			oopsHandler(c, w, r, RespHTML, "Could not upload file: "+err.Error())
			return
		}

		http.Redirect(w, r, Config.sitePath+upload.Filename, 303)
	}
}

func uploadHeaderProcess(r *http.Request, upReq *UploadRequest) {
	if r.Header.Get("Linx-Randomize") == "yes" {
		upReq.randomBarename = true
	}

	upReq.deletionKey = r.Header.Get("Linx-Delete-Key")

	// Get seconds until expiry. Non-integer responses never expire.
	expStr := r.Header.Get("Linx-Expiry")
	upReq.expiry = parseExpiry(expStr)

}

func processUpload(upReq UploadRequest) (upload Upload, err error) {
	// Determine the appropriate filename, then write to disk
	barename, extension := barePlusExt(upReq.filename)

	if upReq.randomBarename || len(barename) == 0 {
		barename = generateBarename()
	}

	var header []byte
	if len(extension) == 0 {
		// Pull the first 512 bytes off for use in MIME detection
		header = make([]byte, 512)
		n, _ := upReq.src.Read(header)
		if n == 0 {
			return upload, errors.New("Empty file")
		}
		header = header[:n]

		// Determine the type of file from header
		kind, err := filetype.Match(header)
		if err != nil || kind.Extension == "unknown" {
			extension = "file"
		} else {
			extension = kind.Extension
		}
	}

	upload.Filename = strings.Join([]string{barename, extension}, ".")
	upload.Filename = strings.Replace(upload.Filename, " ", "", -1)

	fileexists, _ := fileBackend.Exists(upload.Filename)

	// Check if the delete key matches, in which case overwrite
	if fileexists {
		metad, merr := metadataRead(upload.Filename)
		if merr == nil {
			if upReq.deletionKey == metad.DeleteKey {
				fileexists = false
			}
		}
	}

	for fileexists {
		counter, err := strconv.Atoi(string(barename[len(barename)-1]))
		if err != nil {
			barename = barename + "1"
		} else {
			barename = barename[:len(barename)-1] + strconv.Itoa(counter+1)
		}
		upload.Filename = strings.Join([]string{barename, extension}, ".")

		fileexists, err = fileBackend.Exists(upload.Filename)
	}

	if fileBlacklist[strings.ToLower(upload.Filename)] {
		return upload, errors.New("Prohibited filename")
	}

	// Get the rest of the metadata needed for storage
	var fileExpiry time.Time
	if upReq.expiry == 0 {
		fileExpiry = expiry.NeverExpire
	} else {
		fileExpiry = time.Now().Add(upReq.expiry)
	}

	bytes, err := fileBackend.Put(upload.Filename, io.MultiReader(bytes.NewReader(header), upReq.src))
	if err != nil {
		return upload, err
	} else if bytes > Config.maxSize {
		fileBackend.Delete(upload.Filename)
		return upload, errors.New("File too large")
	}

	upload.Metadata, err = generateMetadata(upload.Filename, fileExpiry, upReq.deletionKey)
	if err != nil {
		fileBackend.Delete(upload.Filename)
		return
	}
	err = metadataWrite(upload.Filename, &upload.Metadata)
	if err != nil {
		fileBackend.Delete(upload.Filename)
		return
	}
	return
}

func generateBarename() string {
	return uniuri.NewLenChars(8, []byte("abcdefghijklmnopqrstuvwxyz0123456789"))
}

func generateJSONresponse(upload Upload, r *http.Request) []byte {
	js, _ := json.Marshal(map[string]string{
		"url":        getSiteURL(r) + upload.Filename,
		"filename":   upload.Filename,
		"delete_key": upload.Metadata.DeleteKey,
		"expiry":     strconv.FormatInt(upload.Metadata.Expiry.Unix(), 10),
		"size":       strconv.FormatInt(upload.Metadata.Size, 10),
		"mimetype":   upload.Metadata.Mimetype,
		"sha256sum":  upload.Metadata.Sha256sum,
	})

	return js
}

var bareRe = regexp.MustCompile(`[^A-Za-z0-9\-]`)
var extRe = regexp.MustCompile(`[^A-Za-z0-9\-\.]`)
var compressedExts = map[string]bool{
	".bz2": true,
	".gz":  true,
	".xz":  true,
}
var archiveExts = map[string]bool{
	".tar": true,
}

func barePlusExt(filename string) (barename, extension string) {
	filename = strings.TrimSpace(filename)
	filename = strings.ToLower(filename)

	extension = path.Ext(filename)
	barename = filename[:len(filename)-len(extension)]
	if compressedExts[extension] {
		ext2 := path.Ext(barename)
		if archiveExts[ext2] {
			barename = barename[:len(barename)-len(ext2)]
			extension = ext2 + extension
		}
	}

	extension = extRe.ReplaceAllString(extension, "")
	barename = bareRe.ReplaceAllString(barename, "")

	extension = strings.Trim(extension, "-.")
	barename = strings.Trim(barename, "-")

	return
}

func parseExpiry(expStr string) time.Duration {
	if expStr == "" {
		return time.Duration(Config.maxExpiry) * time.Second
	} else {
		fileExpiry, err := strconv.ParseUint(expStr, 10, 64)
		if err != nil {
			return time.Duration(Config.maxExpiry) * time.Second
		} else {
			if Config.maxExpiry > 0 && (fileExpiry > Config.maxExpiry || fileExpiry == 0) {
				fileExpiry = Config.maxExpiry
			}
			return time.Duration(fileExpiry) * time.Second
		}
	}
}
