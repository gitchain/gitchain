//// Name Allocation Transaction (NAT)
package transaction

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/binary"
	"encoding/gob"
)

func init() {
	gob.Register(&NameAllocation{})
}

const (
	NAME_ALLOCATION_VERSION = 1
)

type NameAllocation struct {
	Version   uint32
	Name      string
	Rand      []byte
	Signature []byte
}

func NewNameAllocation(name string, random []byte, privateKey *rsa.PrivateKey) (txn *NameAllocation, err error) {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, sha1.Sum(append([]byte("ALLOCATE"), name...)))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, buf.Bytes())
	return &NameAllocation{
			Version:   NAME_ALLOCATION_VERSION,
			Name:      name,
			Rand:      random,
			Signature: sig},
		err
}

func (txn *NameAllocation) Verify(publicKey *rsa.PublicKey) bool {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, sha1.Sum(append([]byte("ALLOCATE"), txn.Name...)))
	err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA1, buf.Bytes(), txn.Signature)
	return err == nil
}

func (txn *NameAllocation) Encode() ([]byte, error) {
	return encode(txn)
}

func (txn *NameAllocation) Hash() []byte {
	return hash(txn)
}
