package discache

import (
	"github.com/cespare/xxhash/v2"
)

func HashString(s string) uint64 {

	return xxhash.Sum64String(s)
}
