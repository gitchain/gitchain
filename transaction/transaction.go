package transaction

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"encoding/binary"
	"encoding/gob"

	"../types"

	"github.com/conformal/fastsha256"
)

//// Interface

type T interface {
	Encode() ([]byte, error)
	Hash() []byte
}

func init() {
	gob.Register(&NameReservation{})
	gob.Register(&NameAllocation{})
	gob.Register(&NameDeallocation{})
}

func hash(t T) []byte {
	buf := new(bytes.Buffer)
	gob.NewEncoder(buf).Encode(t)
	encoded := buf.Bytes()
	buf.Reset()
	binary.Write(buf, binary.BigEndian, fastsha256.Sum256(encoded))
	return buf.Bytes()
}

func encode(t T) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(&t)
	return buf.Bytes(), err
}

func Decode(b []byte) (T, error) {
	var t T
	buf := bytes.NewBuffer(b)
	enc := gob.NewDecoder(buf)
	err := enc.Decode(&t)
	return t, err
}

//// Name Reservation Transaction (NRT)
const (
	NAME_RESERVATION_VERSION        = 1
	NAME_RESERVATION_TAG     uint16 = 0x0001
)

type NameReservation struct {
	Version   uint32
	Hashed    types.Hash
	PublicKey rsa.PublicKey
}

func NewNameReservation(name string, publicKey *rsa.PublicKey) (txn *NameReservation, random []byte) {
	buf := make([]byte, 4)
	rand.Read(buf)
	return &NameReservation{
			Version:   NAME_RESERVATION_VERSION,
			Hashed:    fastsha256.Sum256(append([]byte(name), buf...)),
			PublicKey: *publicKey},
		buf
}

func (txn *NameReservation) Encode() ([]byte, error) {
	return encode(txn)
}

func (txn *NameReservation) Hash() []byte {
	return hash(txn)
}

//// Name Allocation Transaction (NAT)
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

//// Name Deallocation Transaction (NAT)
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
