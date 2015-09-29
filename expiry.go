package main

import (
	"time"
)

// Get what the unix timestamp will be in "seconds".
func getFutureTimestamp(seconds int64) (ts int64) {
	now := int64(time.Now().Unix())

	if seconds == 0 {
		ts = 0
	} else {
		ts = now + seconds
	}

	return
}

// Determine if a file with expiry set to "ts" has expired yet
func isTsExpired(ts int64) (expired bool) {
	now := int64(time.Now().Unix())

	if ts == 0 {
		expired = false
	} else if now > ts {
		expired = true
	} else {
		expired = false
	}

	return
}

// Determine if the given filename is expired
func isFileExpired(filename string) bool {
	exp, _ := metadataGetExpiry(filename)

	return isTsExpired(exp)
}
