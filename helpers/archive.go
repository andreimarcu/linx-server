package helpers

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"sort"
)

type ReadSeekerAt interface {
	io.Reader
	io.Seeker
	io.ReaderAt
}

func ListArchiveFiles(mimetype string, size int64, r ReadSeekerAt) (files []string, err error) {
	if mimetype == "application/x-tar" {
		tReadr := tar.NewReader(r)
		for {
			hdr, err := tReadr.Next()
			if err == io.EOF || err != nil {
				break
			}
			if hdr.Typeflag == tar.TypeDir || hdr.Typeflag == tar.TypeReg {
				files = append(files, hdr.Name)
			}
		}
		sort.Strings(files)
	} else if mimetype == "application/x-gzip" {
		gzf, err := gzip.NewReader(r)
		if err == nil {
			tReadr := tar.NewReader(gzf)
			for {
				hdr, err := tReadr.Next()
				if err == io.EOF || err != nil {
					break
				}
				if hdr.Typeflag == tar.TypeDir || hdr.Typeflag == tar.TypeReg {
					files = append(files, hdr.Name)
				}
			}
			sort.Strings(files)
		}
	} else if mimetype == "application/x-bzip" {
		bzf := bzip2.NewReader(r)
		tReadr := tar.NewReader(bzf)
		for {
			hdr, err := tReadr.Next()
			if err == io.EOF || err != nil {
				break
			}
			if hdr.Typeflag == tar.TypeDir || hdr.Typeflag == tar.TypeReg {
				files = append(files, hdr.Name)
			}
		}
		sort.Strings(files)
	} else if mimetype == "application/zip" {
		zf, err := zip.NewReader(r, size)
		if err == nil {
			for _, f := range zf.File {
				files = append(files, f.Name)
			}
		}
		sort.Strings(files)
	}

	return
}
