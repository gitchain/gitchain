package types

import (
	"bytes"
	"encoding/hex"
)

type Hash []byte

func EmptyHash() Hash {
	return make([]byte, 32)
}

func (h Hash) String() string {
	return hex.EncodeToString(h)
}

func (h Hash) Equals(h1 Hash) bool {
	return bytes.Compare(h, h) == 0
}
