package localfs

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/andreimarcu/linx-server/torrent"
)

type LocalfsBackend struct {
	basePath string
}

func (b LocalfsBackend) Delete(key string) error {
	return os.Remove(path.Join(b.basePath, key))
}

func (b LocalfsBackend) Exists(key string) (bool, error) {
	_, err := os.Stat(path.Join(b.basePath, key))
	return err == nil, err
}

func (b LocalfsBackend) Get(key string) ([]byte, error) {
	return ioutil.ReadFile(path.Join(b.basePath, key))
}

func (b LocalfsBackend) Put(key string, r io.Reader) (int64, error) {
	dst, err := os.Create(path.Join(b.basePath, key))
	if err != nil {
		return 0, err
	}
	defer dst.Close()

	bytes, err := io.Copy(dst, r)
	if bytes == 0 {
		b.Delete(key)
		return bytes, errors.New("Empty file")
	} else if err != nil {
		b.Delete(key)
		return bytes, err
	}

	return bytes, err
}

func (b LocalfsBackend) Open(key string) (backends.ReadSeekCloser, error) {
	return os.Open(path.Join(b.basePath, key))
}

func (b LocalfsBackend) ServeFile(key string, w http.ResponseWriter, r *http.Request) error {
	filePath := path.Join(b.basePath, key)
	http.ServeFile(w, r, filePath)
	return nil
}

func (b LocalfsBackend) Size(key string) (int64, error) {
	fileInfo, err := os.Stat(path.Join(b.basePath, key))
	if err != nil {
		return 0, err
	}

	return fileInfo.Size(), nil
}

func (b LocalfsBackend) GetTorrent(fileName string, url string) (t torrent.Torrent, err error) {
	chunk := make([]byte, torrent.TORRENT_PIECE_LENGTH)

	t = torrent.Torrent{
		Encoding: "UTF-8",
		Info: torrent.TorrentInfo{
			PieceLength: torrent.TORRENT_PIECE_LENGTH,
			Name:        fileName,
		},
		UrlList: []string{url},
	}

	f, err := b.Open(fileName)
	if err != nil {
		return
	}
	defer f.Close()

	for {
		n, err := f.Read(chunk)
		if err == io.EOF {
			break
		} else if err != nil {
			return t, err
		}

		t.Info.Length += n
		t.Info.Pieces += string(torrent.HashPiece(chunk[:n]))
	}

	return
}

func (b LocalfsBackend) List() ([]string, error) {
	var output []string

	files, err := ioutil.ReadDir(b.basePath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		output = append(output, file.Name())
	}

	return output, nil
}

func NewLocalfsBackend(basePath string) LocalfsBackend {
	return LocalfsBackend{basePath: basePath}
}
