package util

import "github.com/conformal/fastsha256"

func SHA256(buf []byte) []byte {
	result := fastsha256.Sum256(buf)
	return result[:]
}
