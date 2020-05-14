package main

import (
	"flag"

	"github.com/andreimarcu/linx-server/cleanup"
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

	cleanup.Cleanup(filesDir, metaDir, noLogs)
}
