package apikeys

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/crypto/scrypt"

	"github.com/zenazn/goji/web"
)

const (
	scryptSalt   = "linx-server"
	scryptN      = 16384
	scryptr      = 8
	scryptp      = 1
	scryptKeyLen = 32
)

type AuthOptions struct {
	AuthFile      string
	UnauthMethods []string
	BasicAuth     bool
	SiteName      string
	SitePath      string
}

type ApiKeysMiddleware struct {
	successHandler http.Handler
	authKeys       []string
	o              AuthOptions
}

func ReadAuthKeys(authFile string) []string {
	var authKeys []string

	f, err := os.Open(authFile)
	if err != nil {
		log.Fatal("Failed to open authfile: ", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		authKeys = append(authKeys, scanner.Text())
	}

	err = scanner.Err()
	if err != nil {
		log.Fatal("Scanner error while reading authfile: ", err)
	}

	return authKeys
}

func CheckAuth(authKeys []string, key string) (result bool, err error) {
	checkKey, err := scrypt.Key([]byte(key), []byte(scryptSalt), scryptN, scryptr, scryptp, scryptKeyLen)
	if err != nil {
		return
	}

	encodedKey := base64.StdEncoding.EncodeToString(checkKey)
	for _, v := range authKeys {
		if encodedKey == v {
			result = true
			return
		}
	}

	result = false
	return
}

func (a ApiKeysMiddleware) getSitePrefix() string {
	prefix := a.o.SitePath
	if len(prefix) <= 0 || prefix[0] != '/' {
		prefix = "/" + prefix
	}
	return prefix
}

func (a ApiKeysMiddleware) goodAuthorizationHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Location", a.getSitePrefix())
	w.WriteHeader(http.StatusFound)
}

func (a ApiKeysMiddleware) badAuthorizationHandler(w http.ResponseWriter, r *http.Request) {
	if a.o.BasicAuth {
		rs := ""
		if a.o.SiteName != "" {
			rs = fmt.Sprintf(` realm="%s"`, a.o.SiteName)
		}
		w.Header().Set("WWW-Authenticate", `Basic`+rs)
	}
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func (a ApiKeysMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var successHandler http.Handler
	prefix := a.getSitePrefix()

	if r.URL.Path == prefix+"auth" {
		successHandler = http.HandlerFunc(a.goodAuthorizationHandler)
	} else {
		successHandler = a.successHandler
	}

	if sliceContains(a.o.UnauthMethods, r.Method) && r.URL.Path != prefix+"auth" {
		// allow unauthenticated methods
		successHandler.ServeHTTP(w, r)
		return
	}

	key := r.Header.Get("Linx-Api-Key")
	if key == "" && a.o.BasicAuth {
		_, password, ok := r.BasicAuth()
		if ok {
			key = password
		}
	}

	result, err := CheckAuth(a.authKeys, key)
	if err != nil || !result {
		http.HandlerFunc(a.badAuthorizationHandler).ServeHTTP(w, r)
		return
	}

	successHandler.ServeHTTP(w, r)
}

func NewApiKeysMiddleware(o AuthOptions) func(*web.C, http.Handler) http.Handler {
	fn := func(c *web.C, h http.Handler) http.Handler {
		return ApiKeysMiddleware{
			successHandler: h,
			authKeys:       ReadAuthKeys(o.AuthFile),
			o:              o,
		}
	}
	return fn
}

func sliceContains(slice []string, s string) bool {
	for _, v := range slice {
		if s == v {
			return true
		}
	}

	return false
}
