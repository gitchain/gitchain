package block

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"time"

	trans "github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/types"
	"github.com/conformal/fastsha256"
	"github.com/xsleonard/go-merkle"
)

const (
	HIGHEST_TARGET = 0x1d00ffff
)

const (
	BLOCK_TAG uint16 = 0xff00
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
	Transactions      []trans.T
}

func init() {
	gob.Register(&Block{})
}

func (b *Block) Hash() []byte {
	buf := bytes.NewBuffer([]byte{})
	buf.Grow(192)
	buf1 := bytes.NewBuffer([]byte{})
	buf1.Grow(32)
	binary.Write(buf, binary.LittleEndian, b.PreviousBlockHash)
	binary.Write(buf, binary.LittleEndian, b.MerkleRootHash)
	binary.Write(buf, binary.LittleEndian, b.Version)
	binary.Write(buf, binary.LittleEndian, b.Timestamp)
	binary.Write(buf, binary.LittleEndian, b.Bits)
	binary.Write(buf, binary.LittleEndian, b.Nonce)
	hash := fastsha256.Sum256(buf.Bytes())
	binary.Write(buf1, binary.BigEndian, hash)
	hash = fastsha256.Sum256(buf1.Bytes())
	buf1.Reset()
	binary.Write(buf1, binary.BigEndian, hash)
	return buf1.Bytes()
}

func (b *Block) MarshalJSON() ([]byte, error) {
	return json.Marshal(hex.EncodeToString(b.Hash()))
}

func NewBlock(previousBlockHash types.Hash, bits uint32, transactions []trans.T) (*Block, error) {
	encodedTransactions := make([][]byte, len(transactions))
	for i := range transactions {
		t, _ := transactions[i].Encode()
		encodedTransactions[i] = make([]byte, len(t))
		copy(encodedTransactions[i], t)
	}
	merkleRootHash, err := merkleRoot(encodedTransactions)
	if err != nil {
		return nil, err
	}
	return &Block{
		Version:           BLOCK_VERSION,
		PreviousBlockHash: previousBlockHash,
		MerkleRootHash:    merkleRootHash,
		Timestamp:         time.Now().UTC().Unix(),
		Bits:              bits,
		Nonce:             0,
		Transactions:      transactions}, nil
}

func (b *Block) Encode() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(&b)
	return buf.Bytes(), err
}

func Decode(encoded []byte) (*Block, error) {
	var blk Block
	buf := bytes.NewBuffer(encoded)
	enc := gob.NewDecoder(buf)
	err := enc.Decode(&blk)
	return &blk, err
}

func merkleRoot(data [][]byte) (types.Hash, error) {
	tree := merkle.NewTree()
	err := tree.Generate(data, fastsha256.New())
	if err != nil {
		return types.EmptyHash(), err
	}
	return types.NewHash(tree.Root().Hash), err
}
