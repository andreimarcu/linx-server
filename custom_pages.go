package main

import (
	"io/ioutil"
	"log"
	"path"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/russross/blackfriday"
)

func initializeCustomPages(customPagesDir string) {
	files, err := ioutil.ReadDir(customPagesDir)
	if err != nil {
		log.Fatal("Error reading the custom pages directory: ", err)
	}

	for _, file := range files {
		fileName := file.Name()

		if len(fileName) <= 3 {
			continue
		}

		if strings.EqualFold(string(fileName[len(fileName)-3:len(fileName)]), ".md") {
			contents, err := ioutil.ReadFile(path.Join(customPagesDir, fileName))
			if err != nil {
				log.Fatalf("Error reading file %s", fileName)
			}

			unsafe := blackfriday.MarkdownCommon(contents)
			html := bluemonday.UGCPolicy().SanitizeBytes(unsafe)

			fileName := fileName[0 : len(fileName)-3]
			customPages[fileName] = string(html)
			customPagesNames[fileName] = strings.ReplaceAll(fileName, "_", " ")
		}
	}
}
