package transaction

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/binary"
)

//// Name Deallocation Transaction (NDT)
const (
	NAME_DEALLOCATION_VERSION = 1
)

type NameDeallocation struct {
	Version   uint32
	Name      string
	Signature []byte
}

func NewNameDeallocation(name string, privateKey *rsa.PrivateKey) (txn *NameDeallocation, err error) {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, sha1.Sum(append([]byte("DEALLOCATE"), name...)))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, buf.Bytes())
	return &NameDeallocation{
			Version:   NAME_DEALLOCATION_VERSION,
			Name:      name,
			Signature: sig},
		err
}

func (txn *NameDeallocation) Verify(publicKey *rsa.PublicKey) bool {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, sha1.Sum(append([]byte("DEALLOCATE"), txn.Name...)))
	err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA1, buf.Bytes(), txn.Signature)
	return err == nil
}

func (txn *NameDeallocation) Encode() ([]byte, error) {
	return encode(txn)
}

func (txn *NameDeallocation) Hash() []byte {
	return hash(txn)
}
