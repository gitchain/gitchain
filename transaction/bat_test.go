package transaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAttribution(t *testing.T) {
	txn, err := NewBlockAttribution()

	if err != nil {
		t.Errorf("error while creating name allocation transaction: %v", err)
	}

	assert.True(t, txn.Valid())
	txn1 := txn
	txn1.Version = 100
	assert.False(t, txn1.Valid())

}

func TestAttributionEncodingDecoding(t *testing.T) {
	txn, _ := NewBlockAttribution()
	testTransactionEncodingDecoding(t, txn)
}
