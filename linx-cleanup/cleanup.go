package main

import (
	"flag"
	"log"

	"github.com/andreimarcu/linx-server/backends/localfs"
	"github.com/andreimarcu/linx-server/backends/metajson"
	"github.com/andreimarcu/linx-server/expiry"
)

func main() {
	var filesDir string
	var metaDir string
	var noLogs bool

	flag.StringVar(&filesDir, "filespath", "files/",
		"path to files directory")
	flag.StringVar(&metaDir, "metapath", "meta/",
		"path to metadata directory")
	flag.BoolVar(&noLogs, "nologs", false,
		"don't log deleted files")
	flag.Parse()

	metaStorageBackend := localfs.NewLocalfsBackend(metaDir)
	metaBackend := metajson.NewMetaJSONBackend(metaStorageBackend)
	fileBackend := localfs.NewLocalfsBackend(filesDir)

	files, err := metaStorageBackend.List()
	if err != nil {
		panic(err)
	}

	for _, filename := range files {
		metadata, err := metaBackend.Get(filename)
		if err != nil {
			if !noLogs {
				log.Printf("Failed to find metadata for %s", filename)
			}
		}

		if expiry.IsTsExpired(metadata.Expiry) {
			if !noLogs {
				log.Printf("Delete %s", filename)
			}
			fileBackend.Delete(filename)
			metaStorageBackend.Delete(filename)
		}
	}
}
