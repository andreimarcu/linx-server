package cleanup

import (
	"log"
	"time"

	"github.com/andreimarcu/linx-server/backends/localfs"
	"github.com/andreimarcu/linx-server/expiry"
)

func Cleanup(filesDir string, metaDir string, noLogs bool) {
	fileBackend := localfs.NewLocalfsBackend(metaDir, filesDir)

	files, err := fileBackend.List()
	if err != nil {
		panic(err)
	}

	for _, filename := range files {
		metadata, err := fileBackend.Head(filename)
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
		}
	}
}

func PeriodicCleanup(minutes time.Duration, filesDir string, metaDir string, noLogs bool) {
	c := time.Tick(minutes)
	for range c {
		Cleanup(filesDir, metaDir, noLogs)
	}

}
