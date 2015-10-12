package main

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"

	"golang.org/x/crypto/scrypt"
)

const (
	scryptSalt   = "linx-server"
	scryptN      = 16384
	scryptr      = 8
	scryptp      = 1
	scryptKeyLen = 32
)

func main() {
	fmt.Printf("Enter key to hash: ")

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()

	checkKey, err := scrypt.Key([]byte(scanner.Text()), []byte(scryptSalt), scryptN, scryptr, scryptp, scryptKeyLen)
	if err != nil {
		return
	}

	fmt.Println(base64.StdEncoding.EncodeToString(checkKey))
}
