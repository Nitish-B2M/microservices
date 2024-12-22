package utils

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	expRand "golang.org/x/exp/rand"
)

func GenerateRandomToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func GenerateRandomID() int {
	r := expRand.New(expRand.NewSource(uint64(time.Now().UnixNano())))
	return int(r.Int31())
}
