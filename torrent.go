package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"time"

	"github.com/zeebo/bencode"
	"github.com/zenazn/goji/web"
)

const (
	TORRENT_PIECE_LENGTH = 262144
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type TorrentInfo struct {
	PieceLength int    `bencode:"piece length"`
	Pieces      []byte `bencode:"pieces"`
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
}

type Torrent struct {
	Encoding string      `bencode:"encoding"`
	Info     TorrentInfo `bencode:"info"`
	UrlList  []string    `bencode:"url-list"`
}

func CreateTorrent(fileName string, filePath string) []byte {
	chunk := make([]byte, TORRENT_PIECE_LENGTH)
	var pieces []byte
	length := 0

	f, err := os.Open(filePath)
	check(err)

	for {
		n, err := f.Read(chunk)
		if err == io.EOF {
			break
		}
		check(err)

		length += n

		h := sha1.New()
		h.Write(chunk)
		pieces = append(pieces, h.Sum(nil)...)
	}

	f.Close()

	torrent := &Torrent{
		Encoding: "UTF-8",
		Info: TorrentInfo{
			PieceLength: TORRENT_PIECE_LENGTH,
			Pieces:      pieces,
			Name:        fileName,
			Length:      length,
		},
		UrlList: []string{fmt.Sprintf("%sselif/%s", Config.siteURL, fileName)},
	}

	data, err := bencode.EncodeBytes(torrent)
	check(err)

	return data
}

func fileTorrentHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fileName := c.URLParams["name"]
	filePath := path.Join(Config.filesDir, fileName)

	if !fileExistsAndNotExpired(fileName) {
		notFoundHandler(c, w, r)
		return
	}

	encoded := CreateTorrent(fileName, filePath)

	w.Header().Set(`Content-Disposition`, fmt.Sprintf(`attachment; filename="%s.torrent"`, fileName))
	http.ServeContent(w, r, "", time.Now(), bytes.NewReader(encoded))
}

// vim:set ts=8 sw=8 noet:
