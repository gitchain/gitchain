package block

import (
	"crypto/ecdsa"
	"testing"

	"github.com/gitchain/gitchain/keys"
	trans "github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/types"
	"github.com/stretchr/testify/assert"
)

func TestNewBlock(t *testing.T) {
	privateKey := generateKey(t)
	txn1, rand := trans.NewNameReservation("my-new-repository")
	txn1e := trans.NewEnvelope(types.EmptyHash(), txn1)
	txn2, _ := trans.NewNameAllocation("my-new-repository", rand)
	txn2e := trans.NewEnvelope(types.EmptyHash(), txn2)
	txn3, _ := trans.NewNameDeallocation("my-new-repository")
	txn3e := trans.NewEnvelope(types.EmptyHash(), txn3)

	txn1e.Sign(privateKey)
	txn2e.Sign(privateKey)
	txn3e.Sign(privateKey)

	transactions := []*trans.Envelope{txn1e, txn2e, txn3e}
	block1, err := NewBlock(types.EmptyHash(), HIGHEST_TARGET, transactions)

	if err != nil {
		t.Errorf("can't create a block because of %v", err)
	}

	assert.Equal(t, transactions, block1.Transactions)
	assert.NotEqual(t, block1.MerkleRootHash, types.EmptyHash())
}

func TestNewBlockSingleTx(t *testing.T) {
	privateKey := generateKey(t)
	txn1, _ := trans.NewNameReservation("my-new-repository")
	txn1e := trans.NewEnvelope(types.EmptyHash(), txn1)
	txn1e.Sign(privateKey)

	transactions := []*trans.Envelope{txn1e}
	block1, err := NewBlock(types.EmptyHash(), HIGHEST_TARGET, transactions)

	if err != nil {
		t.Errorf("can't create a block because of %v", err)
		return
	}

	assert.Equal(t, transactions, block1.Transactions)
	assert.NotEqual(t, block1.MerkleRootHash, types.EmptyHash())
}

func TestNewBlockNoTx(t *testing.T) {
	transactions := []*trans.Envelope{}
	block1, err := NewBlock(types.EmptyHash(), HIGHEST_TARGET, transactions)

	if err != nil {
		t.Errorf("can't create a block because of %v", err)
	}

	assert.Equal(t, transactions, block1.Transactions)
	assert.Equal(t, block1.MerkleRootHash, types.EmptyHash())
}

func TestEncodeDecode(t *testing.T) {
	privateKey := generateKey(t)
	txn1, rand := trans.NewNameReservation("my-new-repository")
	txn1e := trans.NewEnvelope(types.EmptyHash(), txn1)
	txn2, _ := trans.NewNameAllocation("my-new-repository", rand)
	txn2e := trans.NewEnvelope(types.EmptyHash(), txn2)
	txn3, _ := trans.NewNameDeallocation("my-new-repository")
	txn3e := trans.NewEnvelope(types.EmptyHash(), txn3)

	txn1e.Sign(privateKey)
	txn2e.Sign(privateKey)
	txn3e.Sign(privateKey)

	transactions := []*trans.Envelope{txn1e, txn2e, txn3e}
	block, err := NewBlock(types.EmptyHash(), HIGHEST_TARGET, transactions)
	if err != nil {
		t.Errorf("can't create a block because of %v", err)
	}

	enc, err := block.Encode()
	if err != nil {
		t.Errorf("error while encoding block: %v", err)
	}

	block1, err := Decode(enc)
	if err != nil {
		t.Errorf("error while encoding block: %v", err)
	}

	assert.Equal(t, block, block1, "encoded and decoded block should be identical to the original one")
}

func TestEncodeDecodeEmptyBlock(t *testing.T) {
	transactions := []*trans.Envelope{}
	block, err := NewBlock(types.EmptyHash(), HIGHEST_TARGET, transactions)
	if err != nil {
		t.Errorf("can't create a block because of %v", err)
	}

	enc, err := block.Encode()
	if err != nil {
		t.Errorf("error while encoding block: %v", err)
	}

	block1, err := Decode(enc)
	if err != nil {
		t.Errorf("error while encoding block: %v", err)
	}

	assert.Equal(t, block, block1, "encoded and decoded block should be identical to the original one")
}

func TestTargetFromBits(t *testing.T) {
	// 0x0404cb * 2 * *(8 * (0x1b - 3)) = 0x00000000000404CB000000000000000000000000000000000000000000000000
	b := targetFromBits(0x1b0404cb).Bytes()
	assert.Equal(t, append(make([]byte, 32-len(b)), b...), []byte{0, 0, 0, 0, 0, 4, 4, 203, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}

func generateKey(t *testing.T) *ecdsa.PrivateKey {
	privateKey, err := keys.GenerateECDSA()
	if err != nil {
		t.Errorf("failed to generate a key")
	}
	return privateKey
}
