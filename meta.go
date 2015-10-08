package main

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"time"

	"bitbucket.org/taruti/mimemagic"
	"github.com/dchest/uniuri"
)

type MetadataJSON struct {
	DeleteKey    string   `json:"delete_key"`
	Sha256sum    string   `json:"sha256sum"`
	Mimetype     string   `json:"mimetype"`
	Size         int64    `json:"size"`
	Expiry       int64    `json:"expiry"`
	ArchiveFiles []string `json:"archive_files,omitempty"`
}

type Metadata struct {
	DeleteKey    string
	Sha256sum    string
	Mimetype     string
	Size         int64
	Expiry       time.Time
	ArchiveFiles []string
}

var NotFoundErr = errors.New("File not found.")
var BadMetadata = errors.New("Corrupted metadata.")

func generateMetadata(fName string, exp time.Time, delKey string) (m Metadata, err error) {
	file, err := os.Open(path.Join(Config.filesDir, fName))
	fileInfo, err := os.Stat(path.Join(Config.filesDir, fName))
	if err != nil {
		return
	}
	defer file.Close()

	m.Size = fileInfo.Size()
	m.Expiry = exp

	if delKey == "" {
		m.DeleteKey = uniuri.NewLen(30)
	} else {
		m.DeleteKey = delKey
	}

	// Get first 512 bytes for mimetype detection
	header := make([]byte, 512)
	file.Read(header)

	m.Mimetype = mimemagic.Match("", header)

	// Compute the sha256sum
	hasher := sha256.New()
	file.Seek(0, 0)
	_, err = io.Copy(hasher, file)
	if err == nil {
		m.Sha256sum = hex.EncodeToString(hasher.Sum(nil))
	}
	file.Seek(0, 0)

	// If archive, grab list of filenames
	if m.Mimetype == "application/x-tar" {
		tReadr := tar.NewReader(file)
		for {
			hdr, err := tReadr.Next()
			if err == io.EOF || err != nil {
				break
			}
			if hdr.Typeflag == tar.TypeDir || hdr.Typeflag == tar.TypeReg {
				m.ArchiveFiles = append(m.ArchiveFiles, hdr.Name)
			}
		}
		sort.Strings(m.ArchiveFiles)
	} else if m.Mimetype == "application/x-gzip" {
		gzf, err := gzip.NewReader(file)
		if err == nil {
			tReadr := tar.NewReader(gzf)
			for {
				hdr, err := tReadr.Next()
				if err == io.EOF || err != nil {
					break
				}
				if hdr.Typeflag == tar.TypeDir || hdr.Typeflag == tar.TypeReg {
					m.ArchiveFiles = append(m.ArchiveFiles, hdr.Name)
				}
			}
			sort.Strings(m.ArchiveFiles)
		}
	} else if m.Mimetype == "application/x-bzip" {
		bzf := bzip2.NewReader(file)
		tReadr := tar.NewReader(bzf)
		for {
			hdr, err := tReadr.Next()
			if err == io.EOF || err != nil {
				break
			}
			if hdr.Typeflag == tar.TypeDir || hdr.Typeflag == tar.TypeReg {
				m.ArchiveFiles = append(m.ArchiveFiles, hdr.Name)
			}
		}
		sort.Strings(m.ArchiveFiles)
	} else if m.Mimetype == "application/zip" {
		zf, err := zip.NewReader(file, m.Size)
		if err == nil {
			for _, f := range zf.File {
				m.ArchiveFiles = append(m.ArchiveFiles, f.Name)
			}
		}
		sort.Strings(m.ArchiveFiles)
	}

	return
}

func metadataWrite(filename string, metadata *Metadata) error {
	file, err := os.Create(path.Join(Config.metaDir, filename))
	if err != nil {
		return err
	}
	defer file.Close()

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

	_, err = file.Write(byt)
	if err != nil {
		return err
	}

	return nil
}

func metadataRead(filename string) (metadata Metadata, err error) {
	b, err := ioutil.ReadFile(path.Join(Config.metaDir, filename))
	if err != nil {
		// Metadata does not exist, generate one
		newMData, err := generateMetadata(filename, neverExpire, "")
		if err != nil {
			return metadata, err
		}
		metadataWrite(filename, &newMData)
		b, err = ioutil.ReadFile(path.Join(Config.metaDir, filename))
		if err != nil {
			return metadata, BadMetadata
		}
	}

	mjson := MetadataJSON{}

	err = json.Unmarshal(b, &mjson)
	if err != nil {
		return metadata, BadMetadata
	}

	metadata.DeleteKey = mjson.DeleteKey
	metadata.Mimetype = mjson.Mimetype
	metadata.ArchiveFiles = mjson.ArchiveFiles
	metadata.Sha256sum = mjson.Sha256sum
	metadata.Expiry = time.Unix(mjson.Expiry, 0)
	metadata.Size = mjson.Size

	return
}
