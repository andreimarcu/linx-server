package apikeys

import (
	"testing"
)

func TestCheckAuth(t *testing.T) {
	authKeys := []string{
		"vhvZ/PT1jeTbTAJ8JdoxddqFtebSxdVb0vwPlYO+4HM=",
		"vFpNprT9wbHgwAubpvRxYCCpA2FQMAK6hFqPvAGrdZo=",
	}

	if r, err := CheckAuth(authKeys, ""); err != nil && r {
		t.Fatal("Authorization passed for empty key")
	}

	if r, err := CheckAuth(authKeys, "thisisnotvalid"); err != nil && r {
		t.Fatal("Authorization passed for invalid key")
	}

	if r, err := CheckAuth(authKeys, "haPVipRnGJ0QovA9nyqK"); err != nil && !r {
		t.Fatal("Authorization failed for valid key")
	}
}
