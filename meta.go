package main

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"sort"
	"time"
	"unicode"

	"github.com/andreimarcu/linx-server/backends"
	"github.com/andreimarcu/linx-server/expiry"
	"github.com/dchest/uniuri"
	"gopkg.in/h2non/filetype.v1"
)

var NotFoundErr = errors.New("File not found.")

func generateMetadata(fName string, exp time.Time, delKey string) (m backends.Metadata, err error) {
	file, err := fileBackend.Open(fName)
	if err != nil {
		return
	}
	defer file.Close()

	m.Size, err = fileBackend.Size(fName)
	if err != nil {
		return
	}

	m.Expiry = exp

	if delKey == "" {
		m.DeleteKey = uniuri.NewLen(30)
	} else {
		m.DeleteKey = delKey
	}

	// Get first 512 bytes for mimetype detection
	header := make([]byte, 512)
	file.Read(header)

	kind, err := filetype.Match(header)
	if err != nil {
		m.Mimetype = "application/octet-stream"
	} else {
		m.Mimetype = kind.MIME.Value
	}

	if m.Mimetype == "" {
		// Check if the file seems anything like text
		if printable(header) {
			m.Mimetype = "text/plain"
		} else {
			m.Mimetype = "application/octet-stream"
		}
	}

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

func metadataWrite(filename string, metadata *backends.Metadata) error {
	return metaBackend.Put(filename, metadata)
}

func metadataRead(filename string) (metadata backends.Metadata, err error) {
	metadata, err = metaBackend.Get(filename)
	if err != nil {
		// Metadata does not exist, generate one
		newMData, err := generateMetadata(filename, expiry.NeverExpire, "")
		if err != nil {
			return metadata, err
		}
		metadataWrite(filename, &newMData)

		metadata, err = metaBackend.Get(filename)
	}

	return
}

func printable(data []byte) bool {
	for i, b := range data {
		r := rune(b)

		// A null terminator that's not at the beginning of the file
		if r == 0 && i == 0 {
			return false
		} else if r == 0 && i < 0 {
			continue
		}

		if r > unicode.MaxASCII {
			return false
		}

	}

	return true
}
