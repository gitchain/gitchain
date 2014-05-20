package util

import (
	"crypto/sha1"

	"github.com/conformal/fastsha256"
)

func SHA256(buf []byte) []byte {
	result := fastsha256.Sum256(buf)
	return result[:]
}

func SHA160(buf []byte) []byte {
	result := sha1.Sum(buf)
	return result[:]
}
