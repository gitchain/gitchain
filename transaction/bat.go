//// Block Allocation Transaction (BAT)
package transaction

import "encoding/gob"

func init() {
	gob.Register(&BlockAttribution{})
}

//// Block Allocation Transaction (BAT)

const (
	BLOCK_ATTRIBUTION_VERSION = 1
)

type BlockAttribution struct {
	Version uint32
}

func NewBlockAttribution() (*BlockAttribution, error) {
	return &BlockAttribution{
		Version: BLOCK_ATTRIBUTION_VERSION}, nil
}

func (txn *BlockAttribution) Valid() bool {
	return (txn.Version == BLOCK_ATTRIBUTION_VERSION)
}

func (txn *BlockAttribution) Encode() ([]byte, error) {
	return encode(txn)
}

func (txn *BlockAttribution) Hash() []byte {
	return hash(txn)
}
