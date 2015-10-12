package main

import (
	"bufio"
	"encoding/base64"
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
	o              AuthOptions
}

func checkAuth(authFile string, decodedAuth []byte) (result bool, err error) {
	f, err := os.Open(authFile)
	if err != nil {
		return
	}

	checkKey, err := scrypt.Key([]byte(decodedAuth), []byte(scryptSalt), scryptN, scryptr, scryptp, scryptKeyLen)
	if err != nil {
		return
	}

	encodedKey := base64.StdEncoding.EncodeToString(checkKey)

	scanner := bufio.NewScanner(bufio.NewReader(f))
	for scanner.Scan() {
		if encodedKey == scanner.Text() {
			result = true
			return
		}
	}

	result = false
	err = scanner.Err()
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

	result, err := checkAuth(a.o.AuthFile, decodedAuth)
	if err != nil || !result {
		a.failureHandler.ServeHTTP(w, r)
		return
	}

	a.successHandler.ServeHTTP(w, r)
}

func UploadAuth(o AuthOptions) func(http.Handler) http.Handler {
	fn := func(h http.Handler) http.Handler {
		return auth{
			successHandler: h,
			failureHandler: http.HandlerFunc(badAuthorizationHandler),
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
