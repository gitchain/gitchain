package transaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAllocation(t *testing.T) {
	_, rand := NewNameReservation("my-new-repository")
	txn, err := NewNameAllocation("my-new-repository", rand)

	if err != nil {
		t.Errorf("error while creating name allocation transaction: %v", err)
	}

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
	_, rand := NewNameReservation("my-new-repository")
	txn, _ := NewNameAllocation("my-new-repository", rand)

	testTransactionEncodingDecoding(t, txn)
}
