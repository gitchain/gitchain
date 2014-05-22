package types

import (
	"bytes"
	"encoding/hex"
)

type Hash []byte

func EmptyHash() Hash {
	return []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
}

func HashEqual(h1, h2 Hash) bool {
	return bytes.Compare(h1, h2) == 0
}

func (h Hash) String() string {
	return hex.EncodeToString(h)
}
