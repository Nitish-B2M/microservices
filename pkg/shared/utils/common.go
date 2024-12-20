package utils

import (
	"time"

	"golang.org/x/exp/rand"
)

func GenerateRandomID() int {
	r := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	return int(r.Int31())
}
