package block

import (
	"time"

	trans "../transaction"
	"../types"
	"github.com/conformal/fastsha256"
	"github.com/xsleonard/go-merkle"
)

const (
	HIGHEST_TARGET = 0x1d00ffff
)

const (
	BLOCK_VERSION = 1
)

type Block struct {
	Version           uint32
	PreviousBlockHash types.Hash
	MerkleRootHash    types.Hash
	Timestamp         int64
	Bits              uint32
	Nonce             uint32
	Transactions      [][]byte
}

func NewBlock(previousBlockHash types.Hash, bits uint32, transactions []trans.T) *Block {
	encodedTransactions := make([][]byte, len(transactions))
	for i := range transactions {
		t, _ := transactions[i].Encode()
		encodedTransactions[i] = make([]byte, len(t))
		copy(encodedTransactions[i], t)
	}
	merkleRootHash, _ := merkleRoot(encodedTransactions)
	return &Block{
		Version:           BLOCK_VERSION,
		PreviousBlockHash: previousBlockHash,
		MerkleRootHash:    merkleRootHash,
		Timestamp:         time.Now().UTC().Unix(),
		Bits:              bits,
		Nonce:             0,
		Transactions:      encodedTransactions}
}

func merkleRoot(data [][]byte) (types.Hash, error) {
	tree := merkle.NewTree()
	err := tree.Generate(data, fastsha256.New())
	if err != nil {
		return types.EmptyHash(), err
	}
	return types.NewHash(tree.Root().Hash), err
}
