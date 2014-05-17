package transaction

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewDeallocation(t *testing.T) {
	privateKey := generateKey(t)
	txn, err := NewNameDeallocation("my-new-repository", privateKey)

	if err != nil {
		t.Errorf("error while creating name allocation transaction: %v", err)
	}

	assert.True(t, txn.Verify(&privateKey.PublicKey))

	txn.Name = "my-old-repository"
	assert.False(t, txn.Verify(&privateKey.PublicKey))

	assert.True(t, txn.Valid())
	txn1 := txn
	txn1.Version = 100
	assert.False(t, txn1.Valid())
	txn1 = txn
	txn1.Name = ""
	assert.False(t, txn1.Valid())
}

func TestDeallocationEncodingDecoding(t *testing.T) {
	privateKey := generateKey(t)
	txn, _ := NewNameDeallocation("my-new-repository", privateKey)
	testTransactionEncodingDecoding(t, txn)
}
