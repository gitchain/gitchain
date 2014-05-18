//// Name Deallocation Transaction (NDT)
package transaction

import "encoding/gob"

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

func (txn *NameDeallocation) Hash() []byte {
	return hash(txn)
}
