package helpers

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"unicode"

	"gopkg.in/h2non/filetype.v1"
)

func DetectMime(r io.ReadSeeker) (string, error) {
	// Get first 512 bytes for mimetype detection
	header := make([]byte, 512)

	r.Seek(0, 0)
	r.Read(header)
	r.Seek(0, 0)

	kind, err := filetype.Match(header)
	if err != nil {
		return "application/octet-stream", err
	} else if kind.MIME.Value != "" {
		return kind.MIME.Value, nil
	}

	// Check if the file seems anything like text
	if printable(header) {
		return "text/plain", nil
	} else {
		return "application/octet-stream", nil
	}
}

func Sha256sum(r io.ReadSeeker) (string, error) {
	hasher := sha256.New()

	r.Seek(0, 0)
	_, err := io.Copy(hasher, r)
	if err != nil {
		return "", err
	}

	r.Seek(0, 0)

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func printable(data []byte) bool {
	for i, b := range data {
		r := rune(b)

		// A null terminator that's not at the beginning of the file
		if r == 0 && i == 0 {
			return false
		} else if r == 0 && i < 0 {
			continue
		}

		if r > unicode.MaxASCII {
			return false
		}

	}

	return true
}
