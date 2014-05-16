//// Block Allocation Transaction (BAT)
package transaction

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"encoding/gob"
	"math/big"
)

func init() {
	gob.Register(&BlockAttribution{})
}

//// Block Allocation Transaction (BAT)

const (
	BLOCK_ATTRIBUTION_VERSION = 1
)

type BlockAttribution struct {
	Version    uint32
	SignatureR []byte
	SignatureS []byte
}

func NewBlockAttribution(privateKey *ecdsa.PrivateKey) (txn *BlockAttribution, err error) {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, sha1.Sum([]byte("ATTRIBUTE")))
	sigR, sigS, err := ecdsa.Sign(rand.Reader, privateKey, buf.Bytes())
	return &BlockAttribution{
			Version:    NAME_ALLOCATION_VERSION,
			SignatureR: sigR.Bytes(),
			SignatureS: sigS.Bytes()},
		err
}

func (txn *BlockAttribution) Verify(publicKey *ecdsa.PublicKey) bool {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, sha1.Sum([]byte("ATTRIBUTE")))
	r := new(big.Int)
	s := new(big.Int)
	r.SetBytes(txn.SignatureR)
	s.SetBytes(txn.SignatureS)
	return ecdsa.Verify(publicKey, buf.Bytes(), r, s)
}

func (txn *BlockAttribution) Encode() ([]byte, error) {
	return encode(txn)
}

func (txn *BlockAttribution) Hash() []byte {
	return hash(txn)
}
