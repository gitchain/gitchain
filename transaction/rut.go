// Reference Update Transaction (RUT)
package transaction

import (
	"encoding/gob"
	"encoding/hex"
	"encoding/json"

	"github.com/gitchain/gitchain/types"
)

func init() {
	gob.Register(&ReferenceUpdate{})
}

const (
	REFERENCE_UPDATE_VERSION = 1
)

type ReferenceUpdate struct {
	Version    uint32
	Repository string
	Ref        string
	Old        types.Hash
	New        types.Hash
}

func (tx *ReferenceUpdate) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"Type":       "Reference Update Tranasction",
		"Version":    tx.Version,
		"Repository": tx.Repository,
		"Ref":        tx.Ref,
		"Old":        hex.EncodeToString(tx.Old),
		"New":        hex.EncodeToString(tx.New),
	})
}

func NewReferenceUpdate(repository, ref string, old, new types.Hash) *ReferenceUpdate {
	return &ReferenceUpdate{
		Version:    REFERENCE_UPDATE_VERSION,
		Repository: repository,
		Ref:        ref,
		Old:        old,
		New:        new}
}

func (txn *ReferenceUpdate) Valid() bool {
	return (txn.Version == REFERENCE_UPDATE_VERSION && len(txn.Repository) > 0 &&
		len(txn.Ref) > 0)
}

func (txn *ReferenceUpdate) Encode() ([]byte, error) {
	return encode(txn)
}

func (txn *ReferenceUpdate) Hash() []byte {
	return hash(txn)
}
