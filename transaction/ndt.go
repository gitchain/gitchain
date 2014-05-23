//// Name Deallocation Transaction (NDT)
package transaction

import (
	"encoding/gob"
	"encoding/json"
	"fmt"

	"github.com/gitchain/gitchain/types"
)

func init() {
	gob.Register(&NameDeallocation{})
}

const (
	NAME_DEALLOCATION_VERSION = 1
)

type NameDeallocation struct {
	Version uint32
	Name    string
}

func (tx *NameDeallocation) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Type":    "Name Deallocation Tranasction",
		"Version": tx.Version,
		"Name":    tx.Name,
	})
}

func NewNameDeallocation(name string) (txn *NameDeallocation, err error) {
	return &NameDeallocation{
		Version: NAME_DEALLOCATION_VERSION,
		Name:    name}, nil
}

func (txn *NameDeallocation) Valid() bool {
	return (txn.Version == NAME_ALLOCATION_VERSION && len(txn.Name) > 0)
}

func (txn *NameDeallocation) Encode() ([]byte, error) {
	return encode(txn)
}

func (txn *NameDeallocation) Hash() types.Hash {
	return hash(txn)
}

func (txn *NameDeallocation) String() string {
	return fmt.Sprintf("NDT %s", txn.Name)
}
