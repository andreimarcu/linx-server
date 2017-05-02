package expiry

import (
	"time"
)

var NeverExpire = time.Unix(0, 0)

// Determine if a file with expiry set to "ts" has expired yet
func IsTsExpired(ts time.Time) bool {
	now := time.Now()
	return ts != NeverExpire && now.After(ts)
}
