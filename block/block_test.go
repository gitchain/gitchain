package block

import (
	"crypto/rand"
	"crypto/rsa"
	trans "gitchain/transaction"
	"gitchain/types"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBlock(t *testing.T) {
	privateKey := generateKey(t)
	txn1, rand := trans.NewNameReservation("my-new-repository", &privateKey.PublicKey)
	txn2, _ := trans.NewNameAllocation("my-new-repository", rand, privateKey)
	txn3, _ := trans.NewNameDeallocation("my-new-repository", privateKey)

	transactions := []trans.T{txn1, txn2, txn3}
	block1 := NewBlock(types.EmptyHash(), HIGHEST_TARGET, transactions)

	decodedTransactions := make([]trans.T, 3)
	for i := range block1.Transactions {
		decodedTransactions[i], _ = trans.Decode(block1.Transactions[i])
	}

	assert.Equal(t, transactions, decodedTransactions)
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
