package block

import (
	"bytes"
	"encoding/binary"
	"math/big"
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

func (b *Block) FindNonce() {
	target := targetFromBits(b.Bits)
	i := big.NewInt(int64(0))
	var n uint32
	buf := bytes.NewBuffer([]byte{})
	buf.Grow(192)
	buf1 := bytes.NewBuffer([]byte{})
	buf1.Grow(32)
loop:
	for n = 0; n < 4294967295; n++ {
		binary.Write(buf, binary.LittleEndian, b.PreviousBlockHash)
		binary.Write(buf, binary.LittleEndian, b.MerkleRootHash)
		binary.Write(buf, binary.LittleEndian, b.Version)
		binary.Write(buf, binary.LittleEndian, b.Timestamp)
		binary.Write(buf, binary.LittleEndian, b.Bits)
		binary.Write(buf, binary.LittleEndian, n) // current nonce
		hash := fastsha256.Sum256(buf.Bytes())
		binary.Write(buf1, binary.BigEndian, hash)
		hash = fastsha256.Sum256(buf1.Bytes())
		buf1.Reset()
		binary.Write(buf1, binary.BigEndian, hash)
		i.SetBytes(buf1.Bytes())
		buf1.Reset()
		buf.Reset()
		if i.Cmp(target) == -1 {
			b.Nonce = n
			return
		}
	}
	// Update timestamp and restart the process
	b.Timestamp = time.Now().UTC().Unix()
	goto loop
}
