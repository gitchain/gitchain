//// Name Allocation Transaction (NAT)
package transaction

import "encoding/gob"

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

func (txn *NameAllocation) Hash() []byte {
	return hash(txn)
}
