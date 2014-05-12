package block

import (
	"gitchain/transaction"
	"gitchain/types"
	"github.com/conformal/fastsha256"
	"github.com/xsleonard/go-merkle"
)

type timestamp uint32

const (
	BLOCK_VERSION = 1
)

type Block struct {
	Version           uint32
	PreviousBlockHash types.Hash
	MerkleRootHash    types.Hash
	Timestamp         timestamp
	Bits              uint32
	Nonce             uint32
	Transactions      []transaction.Transaction
}

func NewBlock(previousBlockHash types.Hash, transactions []transaction.Transaction) *Block {
	return &Block{
		Version:           BLOCK_VERSION,
		PreviousBlockHash: previousBlockHash,
		Timestamp:         time.Now().UTC().Unix(),
		Transactions:      transactions}
}

func merkleRoot([][]byte) (types.Hash, err) {
	tree := merkle.NewTree()
	err := tree.Generate([][]byte{[]byte("hello"), []byte("world")}, fastsha256.New())
	if err != nil {
		return nil, err
	}
	return tree.Root().Hash, err
}
