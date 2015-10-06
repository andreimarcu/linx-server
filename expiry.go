package main

import (
	"time"
)

// Determine if a file with expiry set to "ts" has expired yet
func isTsExpired(ts time.Time) bool {
	now := time.Now()
	return !ts.IsZero() && now.After(ts)
}

// Determine if the given filename is expired
func isFileExpired(filename string) bool {
	exp, _ := metadataGetExpiry(filename)
	return isTsExpired(exp)
}
