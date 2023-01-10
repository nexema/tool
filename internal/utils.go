package internal

import (
	"crypto/sha256"
	"hash"
)

// hash hashes input using sha256
var hashInst hash.Hash = sha256.New()

func HashString(input string) string {
	hashInst.Reset()
	hashInst.Write([]byte(input))
	bs := hashInst.Sum(nil)
	return string(bs)
}
