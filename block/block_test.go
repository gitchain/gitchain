package block

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	trans "../transaction"
	"../types"
	"github.com/stretchr/testify/assert"
)

func TestNewBlock(t *testing.T) {
	privateKey := generateKey(t)
	txn1, rand := trans.NewNameReservation("my-new-repository", &privateKey.PublicKey)
	txn2, _ := trans.NewNameAllocation("my-new-repository", rand, privateKey)
	txn3, _ := trans.NewNameDeallocation("my-new-repository", privateKey)

	transactions := []trans.T{txn1, txn2, txn3}
	block1 := NewBlock(types.EmptyHash(), HIGHEST_TARGET, transactions)

	assert.Equal(t, transactions, block1.Transactions)
	assert.NotEqual(t, block1.MerkleRootHash, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
}

func TestEncodeDecode(t *testing.T) {
	privateKey := generateKey(t)
	txn1, rand := trans.NewNameReservation("my-new-repository", &privateKey.PublicKey)
	txn2, _ := trans.NewNameAllocation("my-new-repository", rand, privateKey)
	txn3, _ := trans.NewNameDeallocation("my-new-repository", privateKey)

	transactions := []trans.T{txn1, txn2, txn3}
	block := NewBlock(types.EmptyHash(), HIGHEST_TARGET, transactions)

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

func generateKey(t *testing.T) *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Errorf("failed to generate a key")
	}
	return privateKey
}
