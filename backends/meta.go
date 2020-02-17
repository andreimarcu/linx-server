package backends

import (
	"errors"
	"time"
)

type Metadata struct {
	DeleteKey    string
	AccessKey    string
	Sha256sum    string
	Mimetype     string
	Size         int64
	Expiry       time.Time
	ArchiveFiles []string
}

var BadMetadata = errors.New("Corrupted metadata.")
