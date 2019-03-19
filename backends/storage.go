package backends

import (
	"errors"
	"io"
	"time"
)

type StorageBackend interface {
	Delete(key string) error
	Exists(key string) (bool, error)
	Head(key string) (Metadata, error)
	Get(key string) (Metadata, io.ReadCloser, error)
	Put(key string, r io.Reader, expiry time.Time, deleteKey string) (Metadata, error)
	PutMetadata(key string, r io.Reader, expiry time.Time, deleteKey string) (Metadata, error)
	Size(key string) (int64, error)
}

type MetaStorageBackend interface {
	StorageBackend
	List() ([]string, error)
}

var NotFoundErr = errors.New("File not found.")
var FileEmptyError = errors.New("Empty file")
