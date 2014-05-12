package types

type Hash [32]byte

func NewHash(bytes []byte) Hash {
	var result Hash
	for i := range bytes {
		result[i] = bytes[i]
	}
	return result
}
