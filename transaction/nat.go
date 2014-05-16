//// Name Allocation Transaction (NAT)
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
	gob.Register(&NameAllocation{})
}

const (
	NAME_ALLOCATION_VERSION = 1
)

type NameAllocation struct {
	Version    uint32
	Name       string
	Rand       []byte
	SignatureR []byte
	SignatureS []byte
}

func NewNameAllocation(name string, random []byte, privateKey *ecdsa.PrivateKey) (txn *NameAllocation, err error) {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, sha1.Sum(append([]byte("ALLOCATE"), name...)))
	sigR, sigS, err := ecdsa.Sign(rand.Reader, privateKey, buf.Bytes())
	return &NameAllocation{
			Version:    NAME_ALLOCATION_VERSION,
			Name:       name,
			Rand:       random,
			SignatureR: sigR.Bytes(),
			SignatureS: sigS.Bytes()},
		err
}

func (txn *NameAllocation) Verify(publicKey *ecdsa.PublicKey) bool {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, sha1.Sum(append([]byte("ALLOCATE"), txn.Name...)))
	r := new(big.Int)
	s := new(big.Int)
	r.SetBytes(txn.SignatureR)
	s.SetBytes(txn.SignatureS)
	return ecdsa.Verify(publicKey, buf.Bytes(), r, s)
}

func (txn *NameAllocation) Encode() ([]byte, error) {
	return encode(txn)
}

func (txn *NameAllocation) Hash() []byte {
	return hash(txn)
}
