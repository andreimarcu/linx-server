package main

import (
	"fmt"
	"testing"

	"github.com/zeebo/bencode"
)

func TestCreateTorrent(t *testing.T) {
	fileName := "server.go"
	var decoded Torrent

	encoded, err := CreateTorrent(fileName, fileName)
	if err != nil {
		t.Fatal(err)
	}

	bencode.DecodeBytes(encoded, &decoded)

	if decoded.Encoding != "UTF-8" {
		t.Fatalf("Encoding was %s, expected UTF-8", decoded.Encoding)
	}

	if decoded.Info.Name != "server.go" {
		t.Fatalf("Name was %s, expected server.go", decoded.Info.Name)
	}

	if decoded.Info.PieceLength <= 0 {
		t.Fatal("Expected a piece length, got none")
	}

	if len(decoded.Info.Pieces) <= 0 {
		t.Fatal("Expected at least one piece, got none")
	}

	if decoded.Info.Length <= 0 {
		t.Fatal("Length was less than or equal to 0, expected more")
	}

	tracker := fmt.Sprintf("%sselif/%s", Config.siteURL, fileName)
	if decoded.UrlList[0] != tracker {
		t.Fatalf("First entry in URL list was %s, expected %s", decoded.UrlList[0], tracker)
	}
}

func TestCreateTorrentWithImage(t *testing.T) {
	var decoded Torrent

	encoded, err := CreateTorrent("test.jpg", "static/images/404.jpg")
	if err != nil {
		t.Fatal(err)
	}

	bencode.DecodeBytes(encoded, &decoded)

	if decoded.Info.Pieces != "r\x01\x80j\x99\x84\n\xd3dZ;1NX\xec;\x9d$+f" {
		t.Fatal("Torrent pieces did not match expected pieces for image")
	}
}

// vim:set ts=8 sw=8 noet:
