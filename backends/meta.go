package backends

import (
	"errors"
	"time"
)

type MetaBackend interface {
	Get(key string) (Metadata, error)
	Put(key string, metadata *Metadata) error
}

type Metadata struct {
	DeleteKey    string
	Sha256sum    string
	Mimetype     string
	Size         int64
	Expiry       time.Time
	ArchiveFiles []string
	ShortURL     string
}

var BadMetadata = errors.New("Corrupted metadata.")
