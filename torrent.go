package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/zeebo/bencode"
	"github.com/zenazn/goji/web"
)

func createTorrent(fileName string, r *http.Request) ([]byte, error) {
	url := fmt.Sprintf("%sselif/%s", getSiteURL(r), fileName)

	t, err := storageBackend.GetTorrent(fileName, url)
	if err != nil {
		return []byte{}, err
	}

	data, err := bencode.EncodeBytes(&t)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func fileTorrentHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fileName := c.URLParams["name"]

	_, err := checkFile(fileName)
	if err == backends.NotFoundErr {
		notFoundHandler(c, w, r)
		return
	} else if err == backends.BadMetadata {
		oopsHandler(c, w, r, RespAUTO, "Corrupt metadata.")
		return
	}

	encoded, err := createTorrent(fileName, r)
	if err != nil {
		oopsHandler(c, w, r, RespHTML, "Could not create torrent.")
		return
	}

	w.Header().Set(`Content-Disposition`, fmt.Sprintf(`attachment; filename="%s.torrent"`, fileName))
	http.ServeContent(w, r, "", time.Now(), bytes.NewReader(encoded))
}
