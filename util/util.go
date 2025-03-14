package util

import (
	"encoding/hex"
	"math/rand"
	"time"
)

var src = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandStringBytesMaskImprSrc(n int) string {
    b := make([]byte, (n+1)/2) // can be simplified to n/2 if n is always even

    if _, err := src.Read(b); err != nil {
            panic(err)
    }

    return hex.EncodeToString(b)[:n]
}