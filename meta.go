package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"time"
)

// Write metadata from Upload struct to file
func metadataWrite(filename string, upload *Upload) error {
	// Write metadata, overwriting if necessary

	file, err := os.Create(path.Join(Config.metaDir, upload.Filename))
	if err != nil {
		return err
	}

	defer file.Close()

	w := bufio.NewWriter(file)

	fmt.Fprintln(w, upload.Expiry.Unix())
	fmt.Fprintln(w, upload.DeleteKey)

	return w.Flush()
}

// Return list of strings from a filename's metadata source
func metadataRead(filename string) ([]string, error) {
	file, err := os.Open(path.Join(Config.metaDir, filename))

	if err != nil {
		return nil, err
	}

	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func metadataGetExpiry(filename string) (expiry time.Time, err error) {
	metadata, err := metadataRead(filename)

	// XXX in this case it's up to the caller to determine proper behavior
	// for a nonexistant metadata file or broken file

	if err != nil {
		return
	}

	if len(metadata) < 1 {
		err = errors.New("ERR: Metadata file does not contain expiry")
		return
	}

	expirySecs, err := strconv.ParseInt(metadata[0], 10, 64)
	expiry = time.Unix(expirySecs, 0)
	return
}

func metadataGetDeleteKey(filename string) (string, error) {
	metadata, err := metadataRead(filename)

	if len(metadata) < 2 {
		err := errors.New("ERR: Metadata file does not contain deletion key")
		return "", err
	}

	if err != nil {
		return "", err
	} else {
		return metadata[1], err
	}
}
