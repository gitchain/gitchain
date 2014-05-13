package transaction

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/binary"

	"github.com/conformal/fastsha256"
	"github.com/gitchain/gitchain/types"
)

//// Interface

type T interface {
	Id() []byte
	Encode() ([]byte, error)
}

//// Name Reservation Transaction (NRT)
const (
	NAME_RESERVATION_VERSION        = 1
	NAME_RESERVATION_TAG     uint16 = 1
)

type NameReservation struct {
	Version   uint32
	Hash      types.Hash
	PublicKey rsa.PublicKey
}

func NewNameReservation(name string, publicKey *rsa.PublicKey) (txn *NameReservation, random []byte) {
	buf := make([]byte, 4)
	rand.Read(buf)
	return &NameReservation{
			Version:   NAME_RESERVATION_VERSION,
			Hash:      fastsha256.Sum256(append([]byte(name), buf...)),
			PublicKey: *publicKey},
		buf
}

func (txn *NameReservation) Encode() ([]byte, error) {
	pubkey, err := x509.MarshalPKIXPublicKey(&txn.PublicKey)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, NAME_RESERVATION_TAG)
	binary.Write(buf, binary.LittleEndian, txn.Version)

	encoded := append(append(buf.Bytes(), txn.Hash[:]...), pubkey...)

	return encoded, nil
}

func (txn *NameReservation) Id() []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.Grow(32)
	encoded, _ := txn.Encode()
	binary.Write(buf, binary.BigEndian, fastsha256.Sum256(encoded))
	return buf.Bytes()
}

func DecodeNameReservation(encoded []byte) (*NameReservation, error) {
	var version uint32
	binary.Read(bytes.NewReader(encoded[2:]), binary.LittleEndian, &version)
	hash := types.NewHash(encoded[6:38])
	publicKey, err := x509.ParsePKIXPublicKey(encoded[38:])
	if err != nil {
		return nil, err
	}

	decoded := &NameReservation{Version: version, Hash: hash, PublicKey: *publicKey.(*rsa.PublicKey)}
	return decoded, nil
}

//// Name Allocation Transaction (NAT)
const (
	NAME_ALLOCATION_VERSION        = 1
	NAME_ALLOCATION_TAG     uint16 = 2
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
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, NAME_ALLOCATION_TAG)
	binary.Write(buf, binary.LittleEndian, txn.Version)
	binary.Write(buf, binary.LittleEndian, uint32(len(txn.Name)))

	encoded := append(append(append(buf.Bytes(), []byte(txn.Name)...), txn.Rand...), txn.Signature...)

	return encoded, nil
}

func (txn *NameAllocation) Id() []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.Grow(32)
	encoded, _ := txn.Encode()
	binary.Write(buf, binary.BigEndian, fastsha256.Sum256(encoded))
	return buf.Bytes()
}

func DecodeNameAllocation(encoded []byte) (*NameAllocation, error) {
	var version uint32
	var nameLen uint32
	binary.Read(bytes.NewReader(encoded[2:]), binary.LittleEndian, &version)
	binary.Read(bytes.NewReader(encoded[6:10]), binary.LittleEndian, &nameLen)

	name := string(encoded[10 : nameLen+10])
	random := encoded[10+nameLen : 10+nameLen+4]
	signature := encoded[10+4+nameLen:]
	decoded := &NameAllocation{Version: version, Name: name, Rand: random, Signature: signature}
	return decoded, nil
}

//// Name Deallocation Transaction (NAT)
const (
	NAME_DEALLOCATION_VERSION        = 1
	NAME_DEALLOCATION_TAG     uint16 = 3
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
	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, NAME_DEALLOCATION_TAG)
	binary.Write(buf, binary.LittleEndian, txn.Version)
	binary.Write(buf, binary.LittleEndian, uint32(len(txn.Name)))

	encoded := append(append(buf.Bytes(), []byte(txn.Name)...), txn.Signature...)

	return encoded, nil
}

func (txn *NameDeallocation) Id() []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.Grow(32)
	encoded, _ := txn.Encode()
	binary.Write(buf, binary.BigEndian, fastsha256.Sum256(encoded))
	return buf.Bytes()
}

func DecodeNameDeallocation(encoded []byte) (*NameDeallocation, error) {
	var version uint32
	var nameLen uint32
	binary.Read(bytes.NewReader(encoded[2:]), binary.LittleEndian, &version)
	binary.Read(bytes.NewReader(encoded[6:10]), binary.LittleEndian, &nameLen)

	name := string(encoded[10 : nameLen+10])
	signature := encoded[10+nameLen:]
	decoded := &NameDeallocation{Version: version, Name: name, Signature: signature}
	return decoded, nil
}

////////////////////////////////////

func Decode(encoded []byte) (T, error) {
	var tag uint16
	binary.Read(bytes.NewReader(encoded), binary.LittleEndian, &tag)
	switch tag {
	case NAME_RESERVATION_TAG:
		result, err := DecodeNameReservation(encoded)
		return result, err
	case NAME_ALLOCATION_TAG:
		result, err := DecodeNameAllocation(encoded)
		return result, err
	case NAME_DEALLOCATION_TAG:
		result, err := DecodeNameDeallocation(encoded)
		return result, err
	default:
		return nil, nil
	}
}
