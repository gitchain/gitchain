//// Block Allocation Transaction (BAT)
package transaction

import (
	"encoding/gob"
	"encoding/json"

	"github.com/gitchain/gitchain/types"
)

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

func (tx *BlockAttribution) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Type":    "Block Attribution Tranasction",
		"Version": tx.Version,
	})
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

func (txn *BlockAttribution) Hash() types.Hash {
	return hash(txn)
}

func (*BlockAttribution) String() string {
	return "BAT"
}
