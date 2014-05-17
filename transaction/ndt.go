//// Name Deallocation Transaction (NDT)
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
	gob.Register(&NameDeallocation{})
}

const (
	NAME_DEALLOCATION_VERSION = 1
)

type NameDeallocation struct {
	Version    uint32
	Name       string
	SignatureR []byte
	SignatureS []byte
}

func NewNameDeallocation(name string, privateKey *ecdsa.PrivateKey) (txn *NameDeallocation, err error) {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, sha1.Sum(append([]byte("DEALLOCATE"), name...)))
	sigR, sigS, err := ecdsa.Sign(rand.Reader, privateKey, buf.Bytes())
	return &NameDeallocation{
			Version:    NAME_DEALLOCATION_VERSION,
			Name:       name,
			SignatureR: sigR.Bytes(),
			SignatureS: sigS.Bytes()},
		err
}

func (txn *NameDeallocation) Valid() bool {
	return (txn.Version == NAME_ALLOCATION_VERSION && len(txn.Name) > 0)
}

func (txn *NameDeallocation) Verify(publicKey *ecdsa.PublicKey) bool {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, sha1.Sum(append([]byte("DEALLOCATE"), txn.Name...)))
	r := new(big.Int)
	s := new(big.Int)
	r.SetBytes(txn.SignatureR)
	s.SetBytes(txn.SignatureS)
	return ecdsa.Verify(publicKey, buf.Bytes(), r, s)
}

func (txn *NameDeallocation) Encode() ([]byte, error) {
	return encode(txn)
}

func (txn *NameDeallocation) Hash() []byte {
	return hash(txn)
}
