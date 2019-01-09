package backends

import (
	"io"
	"net/http"

	"github.com/andreimarcu/linx-server/torrent"
)

type ReadSeekCloser interface {
	io.Reader
	io.Closer
	io.Seeker
	io.ReaderAt
}

type StorageBackend interface {
	Delete(key string) error
	Exists(key string) (bool, error)
	Get(key string) ([]byte, error)
	Put(key string, r io.Reader) (int64, error)
	Open(key string) (ReadSeekCloser, error)
	ServeFile(key string, w http.ResponseWriter, r *http.Request) error
	Size(key string) (int64, error)
	GetTorrent(fileName string, url string) (torrent.Torrent, error)
}

type MetaStorageBackend interface {
	StorageBackend
	List() ([]string, error)
}
