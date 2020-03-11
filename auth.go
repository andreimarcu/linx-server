package main

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
}

type auth struct {
	successHandler http.Handler
	failureHandler http.Handler
	authKeys       []string
	o              AuthOptions
}

func readAuthKeys(authFile string) []string {
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

func checkAuth(authKeys []string, key string) (result bool, err error) {
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

func (a auth) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if sliceContains(a.o.UnauthMethods, r.Method) {
		// allow unauthenticated methods
		a.successHandler.ServeHTTP(w, r)
		return
	}

	key := r.Header.Get("Linx-Api-Key")
	if key == "" && Config.basicAuth {
		_, password, ok := r.BasicAuth()
		if ok {
			key = password
		}
	}

	result, err := checkAuth(a.authKeys, key)
	if err != nil || !result {
		a.failureHandler.ServeHTTP(w, r)
		return
	}

	a.successHandler.ServeHTTP(w, r)
}

func UploadAuth(o AuthOptions) func(*web.C, http.Handler) http.Handler {
	fn := func(c *web.C, h http.Handler) http.Handler {
		return auth{
			successHandler: h,
			failureHandler: http.HandlerFunc(badAuthorizationHandler),
			authKeys:       readAuthKeys(o.AuthFile),
			o:              o,
		}
	}
	return fn
}

func badAuthorizationHandler(w http.ResponseWriter, r *http.Request) {
	if Config.basicAuth {
		rs := ""
		if Config.siteName != "" {
			rs = fmt.Sprintf(` realm="%s"`, Config.siteName)
		}
		w.Header().Set("WWW-Authenticate", `Basic`+rs)
	}
	http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
}

func sliceContains(slice []string, s string) bool {
	for _, v := range slice {
		if s == v {
			return true
		}
	}

	return false
}
