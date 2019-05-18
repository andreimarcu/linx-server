package helpers

import (
	"bytes"
	"encoding/hex"
	"io"
	"unicode"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/minio/sha256-simd"
	"gopkg.in/h2non/filetype.v1"
)

func GenerateMetadata(r io.Reader) (m backends.Metadata, err error) {
	// Since we don't have the ability to seek within a file, we can use a
	// Buffer in combination with a TeeReader to keep a copy of the bytes
	// we read when detecting the file type. These bytes are still needed
	// to hash the file and determine its size and cannot be discarded.
	var buf bytes.Buffer
	teeReader := io.TeeReader(r, &buf)

	// Get first 512 bytes for mimetype detection
	header := make([]byte, 512)
	_, err = teeReader.Read(header)
	if err != nil {
		return
	}

	// Create a Hash and a MultiReader that includes the Buffer we created
	// above along with the original Reader, which will have the rest of
	// the file.
	hasher := sha256.New()
	multiReader := io.MultiReader(&buf, r)

	// Copy everything into the Hash, then use the number of bytes written
	// as the file size.
	var readLen int64
	readLen, err = io.Copy(hasher, multiReader)
	if err != nil {
		return
	} else {
		m.Size += readLen
	}

	// Get the hex-encoded string version of the Hash checksum
	m.Sha256sum = hex.EncodeToString(hasher.Sum(nil))

	// Use the bytes we extracted earlier and attempt to determine the file
	// type
	kind, err := filetype.Match(header)
	if err != nil {
		m.Mimetype = "application/octet-stream"
		return m, err
	} else if kind.MIME.Value != "" {
		m.Mimetype = kind.MIME.Value
	} else if printable(header) {
		m.Mimetype = "text/plain"
	} else {
		m.Mimetype = "application/octet-stream"
	}

	return
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
