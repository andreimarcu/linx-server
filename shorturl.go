package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/zenazn/goji/web"
)

type shortenerRequest struct {
	LongURL string `json:"longUrl"`
}

type shortenerResponse struct {
	Kind    string `json:"kind"`
	ID      string `json:"id"`
	LongURL string `json:"longUrl"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func shortURLHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fileName := c.URLParams["name"]

	err := checkFile(fileName)
	if err == NotFoundErr {
		notFoundHandler(c, w, r)
		return
	}

	metadata, err := metadataRead(fileName)
	if err != nil {
		oopsHandler(c, w, r, RespJSON, "Corrupt metadata.")
		return
	}

	if metadata.ShortURL == "" {
		url, err := shortenURL(getSiteURL(r) + fileName)
		if err != nil {
			oopsHandler(c, w, r, RespJSON, err.Error())
			return
		}

		metadata.ShortURL = url

		err = metadataWrite(fileName, &metadata)
		if err != nil {
			oopsHandler(c, w, r, RespJSON, "Corrupt metadata.")
			return
		}
	}

	js, _ := json.Marshal(map[string]string{
		"shortUrl": metadata.ShortURL,
	})
	w.Write(js)
	return
}

func shortenURL(url string) (string, error) {
	apiURL := "https://www.googleapis.com/urlshortener/v1/url"
	if Config.googleShorterAPIKey != "" {
		apiURL += "?key=" + Config.googleShorterAPIKey
	}

	jsonStr, _ := json.Marshal(shortenerRequest{LongURL: url})

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	shortenerResponse := new(shortenerResponse)
	err = json.NewDecoder(resp.Body).Decode(shortenerResponse)
	if err != nil {
		return "", err
	}

	if shortenerResponse.Error.Message != "" {
		return "", errors.New(shortenerResponse.Error.Message)
	}

	return shortenerResponse.ID, nil
}
