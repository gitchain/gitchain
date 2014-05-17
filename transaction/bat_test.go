package transaction

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAttribution(t *testing.T) {
	privateKey := generateKey(t)
	txn, err := NewBlockAttribution(privateKey)

	if err != nil {
		t.Errorf("error while creating name allocation transaction: %v", err)
	}

	assert.True(t, txn.Verify(&privateKey.PublicKey))
	txn.SignatureR = []byte("boom")
	assert.False(t, txn.Verify(&privateKey.PublicKey))

	assert.True(t, txn.Valid())
	txn1 := txn
	txn1.Version = 100
	assert.False(t, txn1.Valid())

}

func TestAttributionEncodingDecoding(t *testing.T) {
	privateKey := generateKey(t)
	txn, _ := NewBlockAttribution(privateKey)
	testTransactionEncodingDecoding(t, txn)
}
