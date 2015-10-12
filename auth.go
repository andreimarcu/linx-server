package main

import (
	"bufio"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/crypto/scrypt"
)

const (
	authPrefix   = "Linx "
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

func checkAuth(authKeys []string, decodedAuth []byte) (result bool, err error) {
	checkKey, err := scrypt.Key([]byte(decodedAuth), []byte(scryptSalt), scryptN, scryptr, scryptp, scryptKeyLen)
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

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, authPrefix) {
		a.failureHandler.ServeHTTP(w, r)
		return
	}

	decodedAuth, err := base64.StdEncoding.DecodeString(authHeader[len(authPrefix):])
	if err != nil {
		a.failureHandler.ServeHTTP(w, r)
		return
	}

	result, err := checkAuth(a.authKeys, decodedAuth)
	if err != nil || !result {
		a.failureHandler.ServeHTTP(w, r)
		return
	}

	a.successHandler.ServeHTTP(w, r)
}

func UploadAuth(o AuthOptions) func(http.Handler) http.Handler {
	var authKeys []string

	f, err := os.Open(o.AuthFile)
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

	fn := func(h http.Handler) http.Handler {
		return auth{
			successHandler: h,
			failureHandler: http.HandlerFunc(badAuthorizationHandler),
			authKeys:       authKeys,
			o:              o,
		}
	}
	return fn
}

func badAuthorizationHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusUnauthorized)
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
