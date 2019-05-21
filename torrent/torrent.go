package torrent

import (
	"crypto/sha1"
)

const (
	torrentPieceLength = 262144
)

type torrentInfo struct {
	PieceLength int    `bencode:"piece length"`
	Pieces      string `bencode:"pieces"`
	Name        string `bencode:"name"`
	Length      int    `bencode:"length"`
}

type torrent struct {
	Encoding string      `bencode:"encoding"`
	Info     torrentInfo `bencode:"info"`
	URLList  []string    `bencode:"url-list"`
}

func hashPiece(piece []byte) []byte {
	h := sha1.New()
	h.Write(piece)
	return h.Sum(nil)
}
