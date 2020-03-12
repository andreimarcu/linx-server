package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/andreimarcu/linx-server/torrent"
	"github.com/zeebo/bencode"
)

func TestCreateTorrent(t *testing.T) {
	fileName := "server.go"
	var decoded torrent.Torrent

	f, err := os.Open("server.go")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	encoded, err := createTorrent(fileName, f, nil)
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

	tracker := fmt.Sprintf("%s%s%s", Config.siteURL, Config.selifPath, fileName)
	if decoded.UrlList[0] != tracker {
		t.Fatalf("First entry in URL list was %s, expected %s", decoded.UrlList[0], tracker)
	}
}

func TestCreateTorrentWithImage(t *testing.T) {
	var decoded torrent.Torrent

	f, err := os.Open("static/images/404.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	encoded, err := createTorrent("test.jpg", f, nil)
	if err != nil {
		t.Fatal(err)
	}

	bencode.DecodeBytes(encoded, &decoded)

	if decoded.Info.Pieces != "\xd6\xff\xbf'^)\x85?\xb4.\xb0\xc1|\xa3\x83\xeeX\xf9\xfd\xd7" {
		t.Fatal("Torrent pieces did not match expected pieces for image")
	}
}
