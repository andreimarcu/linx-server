package helpers

import (
	"bytes"
	"strings"
	"testing"
	"unicode/utf16"
)

func TestGenerateMetadata(t *testing.T) {
	r := strings.NewReader("This is my test content")
	m, err := GenerateMetadata(r)
	if err != nil {
		t.Fatal(err)
	}

	expectedSha256sum := "966152d20a77e739716a625373ee15af16e8f4aec631a329a27da41c204b0171"
	if m.Sha256sum != expectedSha256sum {
		t.Fatalf("Sha256sum was %q instead of expected value of %q", m.Sha256sum, expectedSha256sum)
	}

	expectedMimetype := "text/plain; charset=utf-8"
	if m.Mimetype != expectedMimetype {
		t.Fatalf("Mimetype was %q instead of expected value of %q", m.Mimetype, expectedMimetype)
	}

	expectedSize := int64(23)
	if m.Size != expectedSize {
		t.Fatalf("Size was %d instead of expected value of %d", m.Size, expectedSize)
	}
}

func TestTextCharsets(t *testing.T) {
	// verify that different text encodings are detected and passed through
	orig := "This is a text string"
	utf16 := utf16.Encode([]rune(orig))
	utf16LE := make([]byte, len(utf16)*2+2)
	utf16BE := make([]byte, len(utf16)*2+2)
	utf8 := []byte(orig)
	utf16LE[0] = 0xff
	utf16LE[1] = 0xfe
	utf16BE[0] = 0xfe
	utf16BE[1] = 0xff
	for i := 0; i < len(utf16); i++ {
		lsb := utf16[i] & 0xff
		msb := utf16[i] >> 8
		utf16LE[i*2+2] = byte(lsb)
		utf16LE[i*2+3] = byte(msb)
		utf16BE[i*2+2] = byte(msb)
		utf16BE[i*2+3] = byte(lsb)
	}

	testcases := []struct {
		data      []byte
		extension string
		mimetype  string
	}{
		{mimetype: "text/plain; charset=utf-8", data: utf8},
		{mimetype: "text/plain; charset=utf-16le", data: utf16LE},
		{mimetype: "text/plain; charset=utf-16be", data: utf16BE},
	}

	for i, testcase := range testcases {
		r := bytes.NewReader(testcase.data)
		m, err := GenerateMetadata(r)
		if err != nil {
			t.Fatalf("[%d] unexpected error return %v\n", i, err)
		}
		if m.Mimetype != testcase.mimetype {
			t.Errorf("[%d] Expected mimetype '%s', got mimetype '%s'\n", i, testcase.mimetype, m.Mimetype)
		}
	}
}
