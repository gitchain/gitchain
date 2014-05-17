package transaction

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAllocation(t *testing.T) {
	privateKey := generateKey(t)
	_, rand := NewNameReservation("my-new-repository", &privateKey.PublicKey)
	txn, err := NewNameAllocation("my-new-repository", rand, privateKey)

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
	txn1 = txn
	for i := 0; i < 100; i++ {
		if i != 4 {
			txn1.Rand = make([]byte, i)
			assert.False(t, txn1.Valid())
		}
	}
}

func TestAllocationEncodingDecoding(t *testing.T) {
	privateKey := generateKey(t)
	_, rand := NewNameReservation("my-new-repository", &privateKey.PublicKey)
	txn, _ := NewNameAllocation("my-new-repository", rand, privateKey)

	testTransactionEncodingDecoding(t, txn)
}
