//// Block Allocation Transaction (BAT)
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
	gob.Register(&BlockAttribution{})
}

//// Block Allocation Transaction (BAT)

const (
	BLOCK_ATTRIBUTION_VERSION = 1
)

type BlockAttribution struct {
	Version   uint32
	Signature []byte
}

func NewBlockAttribution(privateKey *rsa.PrivateKey) (txn *BlockAttribution, err error) {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, sha1.Sum([]byte("ATTRIBUTE")))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA1, buf.Bytes())
	return &BlockAttribution{
			Version:   NAME_ALLOCATION_VERSION,
			Signature: sig},
		err
}

func (txn *BlockAttribution) Verify(publicKey *rsa.PublicKey) bool {
	buf := bytes.NewBuffer([]byte{})
	binary.Write(buf, binary.LittleEndian, sha1.Sum([]byte("ATTRIBUTE")))
	err := rsa.VerifyPKCS1v15(publicKey, crypto.SHA1, buf.Bytes(), txn.Signature)
	return err == nil
}

func (txn *BlockAttribution) Encode() ([]byte, error) {
	return encode(txn)
}

func (txn *BlockAttribution) Hash() []byte {
	return hash(txn)
}
