package metajson

import (
	"bytes"
	"encoding/json"
	"time"

	"github.com/andreimarcu/linx-server/backends"
)

type MetadataJSON struct {
	DeleteKey    string   `json:"delete_key"`
	Sha256sum    string   `json:"sha256sum"`
	Mimetype     string   `json:"mimetype"`
	Size         int64    `json:"size"`
	Expiry       int64    `json:"expiry"`
	ArchiveFiles []string `json:"archive_files,omitempty"`
}

type MetaJSONBackend struct {
	storage backends.MetaStorageBackend
}

func (m MetaJSONBackend) Put(key string, metadata *backends.Metadata) error {
	mjson := MetadataJSON{}
	mjson.DeleteKey = metadata.DeleteKey
	mjson.Mimetype = metadata.Mimetype
	mjson.ArchiveFiles = metadata.ArchiveFiles
	mjson.Sha256sum = metadata.Sha256sum
	mjson.Expiry = metadata.Expiry.Unix()
	mjson.Size = metadata.Size

	byt, err := json.Marshal(mjson)
	if err != nil {
		return err
	}

	if _, err := m.storage.Put(key, bytes.NewBuffer(byt)); err != nil {
		return err
	}

	return nil
}

func (m MetaJSONBackend) Get(key string) (metadata backends.Metadata, err error) {
	b, err := m.storage.Get(key)
	if err != nil {
		return metadata, backends.BadMetadata
	}

	mjson := MetadataJSON{}

	err = json.Unmarshal(b, &mjson)
	if err != nil {
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

func NewMetaJSONBackend(storage backends.MetaStorageBackend) MetaJSONBackend {
	return MetaJSONBackend{storage: storage}
}
