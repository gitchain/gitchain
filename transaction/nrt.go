//// Name Reservation Transaction (NRT)
package transaction

import (
	"crypto/rand"
	"encoding/gob"

	"github.com/gitchain/gitchain/types"
	"github.com/gitchain/gitchain/util"
)

func init() {
	gob.Register(&NameReservation{})
}

const (
	NAME_RESERVATION_VERSION = 1
)

type NameReservation struct {
	Version uint32
	Hashed  types.Hash
}

func NewNameReservation(name string) (*NameReservation, []byte) {
	buf := make([]byte, 4)
	rand.Read(buf)
	return &NameReservation{
		Version: NAME_RESERVATION_VERSION,
		Hashed:  util.SHA256(append([]byte(name), buf...))}, buf
}

func (txn *NameReservation) Valid() bool {
	return (txn.Version == NAME_RESERVATION_VERSION && len(txn.Hashed) == 32)
}

func (txn *NameReservation) Encode() ([]byte, error) {
	return encode(txn)
}

func (txn *NameReservation) Hash() []byte {
	return hash(txn)
}
