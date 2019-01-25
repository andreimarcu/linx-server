package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/andreimarcu/linx-server/expiry"
	"github.com/andreimarcu/linx-server/torrent"
	"github.com/zeebo/bencode"
	"github.com/zenazn/goji/web"
)

func createTorrent(fileName string, f io.Reader, r *http.Request) ([]byte, error) {
	url := getSiteURL(r) + Config.selifPath + fileName
	chunk := make([]byte, torrent.TORRENT_PIECE_LENGTH)

	t := torrent.Torrent{
		Encoding: "UTF-8",
		Info: torrent.TorrentInfo{
			PieceLength: torrent.TORRENT_PIECE_LENGTH,
			Name:        fileName,
		},
		UrlList: []string{url},
	}

	for {
		n, err := io.ReadFull(f, chunk)
		if err == io.EOF {
			break
		} else if err != nil && err != io.ErrUnexpectedEOF {
			return []byte{}, err
		}

		t.Info.Length += n
		t.Info.Pieces += string(torrent.HashPiece(chunk[:n]))
	}

	data, err := bencode.EncodeBytes(&t)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func fileTorrentHandler(c web.C, w http.ResponseWriter, r *http.Request) {
	fileName := c.URLParams["name"]

	metadata, f, err := storageBackend.Get(fileName)
	if err == backends.NotFoundErr {
		notFoundHandler(c, w, r)
		return
	} else if err == backends.BadMetadata {
		oopsHandler(c, w, r, RespAUTO, "Corrupt metadata.")
		return
	} else if err != nil {
		oopsHandler(c, w, r, RespAUTO, err.Error())
		return
	}
	defer f.Close()

	if expiry.IsTsExpired(metadata.Expiry) {
		storageBackend.Delete(fileName)
		notFoundHandler(c, w, r)
		return
	}

	encoded, err := createTorrent(fileName, f, r)
	if err != nil {
		oopsHandler(c, w, r, RespHTML, "Could not create torrent.")
		return
	}

	w.Header().Set(`Content-Disposition`, fmt.Sprintf(`attachment; filename="%s.torrent"`, fileName))
	http.ServeContent(w, r, "", time.Now(), bytes.NewReader(encoded))
}
