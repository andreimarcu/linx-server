package helpers

import (
	"strings"
	"testing"
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

	expectedMimetype := "text/plain"
	if m.Mimetype != expectedMimetype {
		t.Fatalf("Mimetype was %q instead of expected value of %q", m.Mimetype, expectedMimetype)
	}

	expectedSize := int64(23)
	if m.Size != expectedSize {
		t.Fatalf("Size was %d instead of expected value of %d", m.Size, expectedSize)
	}
}

func TestSVGMimetype(t *testing.T) {
	testcases := []struct {
		mimetype  string
		inputdata string
	}{
		{mimetype: "image/svg+xml", inputdata: "<svg width=\"2042pt\" height=\"810pt\">\nsvg things\n</svg>\n"},
		{mimetype: "text/plain", inputdata: "Not an SVG!\nThis is sooo not an SVG file.\n"},
	}

	for i, record := range testcases {
		r := strings.NewReader(record.inputdata)
		m, err := GenerateMetadata(r)
		if err != nil {
			t.Fatal(err)
		}
		if record.mimetype != m.Mimetype {
			t.Fatalf("[svg testcase %d] Mimetype was %s instead of expected value %s", i, m.Mimetype, record.mimetype)
		}
	}
}
