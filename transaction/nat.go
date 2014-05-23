//// Name Allocation Transaction (NAT)
package transaction

import (
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/gitchain/gitchain/types"
)

func init() {
	gob.Register(&NameAllocation{})
}

const (
	NAME_ALLOCATION_VERSION = 1
)

type NameAllocation struct {
	Version uint32
	Name    string
	Rand    []byte
}

func (tx *NameAllocation) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Type":    "Name Allocation Tranasction",
		"Version": tx.Version,
		"Name":    tx.Name,
		"Rand":    hex.EncodeToString(tx.Rand),
	})
}

func NewNameAllocation(name string, random []byte) (*NameAllocation, error) {
	return &NameAllocation{
			Version: NAME_ALLOCATION_VERSION,
			Name:    name,
			Rand:    random},
		nil
}

func (txn *NameAllocation) Valid() bool {
	return (txn.Version == NAME_ALLOCATION_VERSION && len(txn.Name) > 0 && len(txn.Rand) == 4)
}

func (txn *NameAllocation) Encode() ([]byte, error) {
	return encode(txn)
}

func (txn *NameAllocation) Hash() types.Hash {
	return hash(txn)
}

func (txn *NameAllocation) String() string {
	return fmt.Sprintf("NAT %s %x", txn.Name, txn.Rand)
}
