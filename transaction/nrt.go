//// Name Reservation Transaction (NRT)
package transaction

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/gob"

	"github.com/conformal/fastsha256"
	"github.com/gitchain/gitchain/types"
)

func init() {
	gob.Register(&NameReservation{})
}

const (
	NAME_RESERVATION_VERSION = 1
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
