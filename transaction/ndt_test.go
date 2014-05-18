package transaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewDeallocation(t *testing.T) {
	txn, err := NewNameDeallocation("my-new-repository")

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
}

func TestDeallocationEncodingDecoding(t *testing.T) {
	txn, _ := NewNameDeallocation("my-new-repository")
	testTransactionEncodingDecoding(t, txn)
}
