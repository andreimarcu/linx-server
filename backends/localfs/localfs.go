package localfs

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/andreimarcu/linx-server/helpers"
	"github.com/andreimarcu/linx-server/torrent"
)

type LocalfsBackend struct {
	metaPath  string
	filesPath string
}

type MetadataJSON struct {
	DeleteKey    string   `json:"delete_key"`
	Sha256sum    string   `json:"sha256sum"`
	Mimetype     string   `json:"mimetype"`
	Size         int64    `json:"size"`
	Expiry       int64    `json:"expiry"`
	ArchiveFiles []string `json:"archive_files,omitempty"`
}

func (b LocalfsBackend) Delete(key string) (err error) {
	err = os.Remove(path.Join(b.filesPath, key))
	if err != nil {
		return
	}
	err = os.Remove(path.Join(b.metaPath, key))
	return
}

func (b LocalfsBackend) Exists(key string) (bool, error) {
	_, err := os.Stat(path.Join(b.filesPath, key))
	return err == nil, err
}

func (b LocalfsBackend) Head(key string) (metadata backends.Metadata, err error) {
	f, err := os.Open(path.Join(b.metaPath, key))
	if os.IsNotExist(err) {
		return metadata, backends.NotFoundErr
	} else if err != nil {
		return metadata, backends.BadMetadata
	}
	defer f.Close()

	decoder := json.NewDecoder(f)

	mjson := MetadataJSON{}
	if err := decoder.Decode(&mjson); err != nil {
		return metadata, backends.BadMetadata
	}

	metadata.DeleteKey = mjson.DeleteKey
	metadata.Mimetype = mjson.Mimetype
	metadata.ArchiveFiles = mjson.ArchiveFiles
	metadata.Sha256sum = mjson.Sha256sum
	metadata.Expiry = time.Unix(mjson.Expiry, 0)
	metadata.Size = mjson.Size

	return
}

func (b LocalfsBackend) Get(key string) (metadata backends.Metadata, f io.ReadCloser, err error) {
	metadata, err = b.Head(key)
	if err != nil {
		return
	}

	f, err = os.Open(path.Join(b.filesPath, key))
	if err != nil {
		return
	}

	return
}

func (b LocalfsBackend) writeMetadata(key string, metadata backends.Metadata) error {
	metaPath := path.Join(b.metaPath, key)

	mjson := MetadataJSON{
		DeleteKey: metadata.DeleteKey,
		Mimetype: metadata.Mimetype,
		ArchiveFiles: metadata.ArchiveFiles,
		Sha256sum: metadata.Sha256sum,
		Expiry: metadata.Expiry.Unix(),
		Size: metadata.Size,
	}

	dst, err := os.Create(metaPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	encoder := json.NewEncoder(dst)
	err = encoder.Encode(mjson)
	if err != nil {
		os.Remove(metaPath)
		return err
	}

	return nil
}

func (b LocalfsBackend) Put(key string, r io.Reader, expiry time.Time, deleteKey string) (m backends.Metadata, err error) {
	filePath := path.Join(b.filesPath, key)

	dst, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer dst.Close()

	bytes, err := io.Copy(dst, r)
	if bytes == 0 {
		os.Remove(filePath)
		return m, errors.New("Empty file")
	} else if err != nil {
		os.Remove(filePath)
		return m, err
	}

	m.Expiry = expiry
	m.DeleteKey = deleteKey
	m.Size = bytes
	m.Mimetype, _ = helpers.DetectMime(dst)
	m.Sha256sum, _ = helpers.Sha256sum(dst)
	m.ArchiveFiles, _ = helpers.ListArchiveFiles(m.Mimetype, m.Size, dst)

	err = b.writeMetadata(key, m)
	if err != nil {
		os.Remove(filePath)
		return
	}

	return
}

func (b LocalfsBackend) Size(key string) (int64, error) {
	fileInfo, err := os.Stat(path.Join(b.filesPath, key))
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

	_, f, err := b.Get(fileName)
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

	files, err := ioutil.ReadDir(b.filesPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		output = append(output, file.Name())
	}

	return output, nil
}

func NewLocalfsBackend(metaPath string, filesPath string) LocalfsBackend {
	return LocalfsBackend{
		metaPath:  metaPath,
		filesPath: filesPath,
	}
}
