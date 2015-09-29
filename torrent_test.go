package main

import (
	"fmt"
	"testing"

	"github.com/zeebo/bencode"
)

func TestCreateTorrent(t *testing.T) {
	fileName := "server.go"
	encoded := CreateTorrent(fileName, fileName)
	var decoded Torrent

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
		t.Fatal("First entry in URL list was %s, expected %s", decoded.UrlList[0], tracker)
	}
}

// vim:set ts=8 sw=8 noet:
