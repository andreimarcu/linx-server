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
