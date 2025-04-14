package auth

import (
	"math/rand"
	"time"
)

func randRange(floor, ceil int64) int64 {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return floor + r.Int63n(ceil-floor)

}
