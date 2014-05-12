package block

import "math/big"

func targetFromBits(bits uint32) *big.Int {
	bits3 := bits - (bits>>24)<<24
	bits1 := 8 * (bits>>24 - 3)
	j := big.NewInt(int64(2))
	j.Exp(j, big.NewInt(int64(bits1)), big.NewInt(0))
	j.Mul(big.NewInt(int64(bits3)), j)
	return j
}
