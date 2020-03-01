package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/flosch/pongo2"
	"github.com/zenazn/goji/web"
)

type accessKeySource int

const (
	accessKeySourceNone accessKeySource = iota
	accessKeySourceCookie
	accessKeySourceHeader
	accessKeySourceForm
	accessKeySourceQuery
)

const accessKeyHeaderName = "Linx-Access-Key"
const accessKeyParamName = "access_key"

var (
	errInvalidAccessKey = errors.New("invalid access key")

	cliUserAgentRe = regexp.MustCompile("(?i)(lib)?curl|wget")
)

func checkAccessKey(r *http.Request, metadata *backends.Metadata) (accessKeySource, error) {
	key := metadata.AccessKey
	if key == "" {
		return accessKeySourceNone, nil
	}

	cookieKey, err := r.Cookie(accessKeyHeaderName)
	if err == nil {
		if cookieKey.Value == key {
			return accessKeySourceCookie, nil
		}
		return accessKeySourceCookie, errInvalidAccessKey
	}

	headerKey := r.Header.Get(accessKeyHeaderName)
	if headerKey == key {
		return accessKeySourceHeader, nil
	} else if headerKey != "" {
		return accessKeySourceHeader, errInvalidAccessKey
	}

	formKey := r.PostFormValue(accessKeyParamName)
	if formKey == key {
		return accessKeySourceForm, nil
	} else if formKey != "" {
		return accessKeySourceForm, errInvalidAccessKey
	}

	queryKey := r.URL.Query().Get(accessKeyParamName)
	if queryKey == key {
		return accessKeySourceQuery, nil
	} else if formKey != "" {
		return accessKeySourceQuery, errInvalidAccessKey
	}

	return accessKeySourceNone, errInvalidAccessKey
}

func setAccessKeyCookies(w http.ResponseWriter, siteURL, fileName, value string, expires time.Time) {
	u, err := url.Parse(siteURL)
	if err != nil {
		log.Printf("cant parse siteURL (%v): %v", siteURL, err)
		return
	}

	cookie := http.Cookie{
		Name:     accessKeyHeaderName,
		Value:    value,
		HttpOnly: true,
		Domain:   u.Hostname(),
		Expires:  expires,
	}

	cookie.Path = path.Join(u.Path, fileName)
	http.SetCookie(w, &cookie)

	cookie.Path = path.Join(u.Path, Config.selifPath, fileName)
	http.SetCookie(w, &cookie)
}

func fileAccessHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	if !Config.noDirectAgents && cliUserAgentRe.MatchString(r.Header.Get("User-Agent")) && !strings.EqualFold("application/json", r.Header.Get("Accept")) {
		fileServeHandler(c, w, r)
		return
	}

	fileName := c.URLParams["name"]

	metadata, err := checkFile(fileName)
	if err == backends.NotFoundErr {
		notFoundHandler(c, w, r)
		return
	} else if err != nil {
		oopsHandler(c, w, r, RespAUTO, "Corrupt metadata.")
		return
	}

	if src, err := checkAccessKey(r, &metadata); err != nil {
		// remove invalid cookie
		if src == accessKeySourceCookie {
			setAccessKeyCookies(w, getSiteURL(r), fileName, "", time.Unix(0, 0))
		}

		if strings.EqualFold("application/json", r.Header.Get("Accept")) {
			dec := json.NewEncoder(w)
			_ = dec.Encode(map[string]string{
				"error": errInvalidAccessKey.Error(),
			})

			return
		}

		_ = renderTemplate(Templates["access.html"], pongo2.Context{
			"filename":   fileName,
			"accesspath": fileName,
		}, r, w)

		return
	}

	if metadata.AccessKey != "" {
		var expiry time.Time
		if Config.accessKeyCookieExpiry != 0 {
			expiry = time.Now().Add(time.Duration(Config.accessKeyCookieExpiry) * time.Second)
		}
		setAccessKeyCookies(w, getSiteURL(r), fileName, metadata.AccessKey, expiry)
	}

	fileDisplayHandler(c, w, r, fileName, metadata)
}
