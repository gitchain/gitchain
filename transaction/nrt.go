//// Name Reservation Transaction (NRT)
package transaction

import (
	"crypto/rand"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"

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

func (tx *NameReservation) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Type":    "Name Reservation Tranasction",
		"Version": tx.Version,
		"Hashed":  hex.EncodeToString(tx.Hashed),
	})
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

func (txn *NameReservation) Hash() types.Hash {
	return hash(txn)
}

func (txn *NameReservation) String() string {
	return fmt.Sprintf("NRT %s", txn.Hashed)
}
