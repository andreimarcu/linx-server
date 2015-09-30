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

type TorrentInfo struct {
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
}

type Torrent struct {
	Encoding string      `bencode:"encoding"`
	Info     TorrentInfo `bencode:"info"`
	UrlList  []string    `bencode:"url-list"`
}

func hashPiece(piece []byte) []byte {
	h := sha1.New()
	h.Write(piece)
	return h.Sum(nil)
}

func CreateTorrent(fileName string, filePath string) ([]byte, error) {
	chunk := make([]byte, TORRENT_PIECE_LENGTH)

	torrent := Torrent{
		Encoding: "UTF-8",
		Info: TorrentInfo{
			PieceLength: TORRENT_PIECE_LENGTH,
			Name:        fileName,
		},
		UrlList: []string{fmt.Sprintf("%sselif/%s", Config.siteURL, fileName)},
	}

	f, err := os.Open(filePath)
	if err != nil {
		return []byte{}, err
	}

	for {
		n, err := f.Read(chunk)
		if err == io.EOF {
			break
		} else if err != nil {
			return []byte{}, err
		}

		torrent.Info.Length += n
		torrent.Info.Pieces += string(hashPiece(chunk[:n]))
	}

	f.Close()

	data, err := bencode.EncodeBytes(&torrent)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func fileTorrentHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fileName := c.URLParams["name"]
	filePath := path.Join(Config.filesDir, fileName)

	if !fileExistsAndNotExpired(fileName) {
		notFoundHandler(c, w, r)
		return
	}

	encoded, err := CreateTorrent(fileName, filePath)
	if err != nil {
		oopsHandler(c, w, r) // 500 - creating torrent failed
		return
	}

	w.Header().Set(`Content-Disposition`, fmt.Sprintf(`attachment; filename="%s.torrent"`, fileName))
	http.ServeContent(w, r, "", time.Now(), bytes.NewReader(encoded))
}

// vim:set ts=8 sw=8 noet:
