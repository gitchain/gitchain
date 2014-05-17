package transaction

import (
	"crypto/ecdsa"
	"testing"

	"github.com/gitchain/gitchain/keys"
	"github.com/stretchr/testify/assert"
)

func testTransactionEncodingDecoding(t *testing.T, txn T) {
	enc, err := txn.Encode()
	if err != nil {
		t.Errorf("error while encoding transaction: %v", err)
	}

	txn1, err := Decode(enc)
	if err != nil {
		t.Errorf("error while encoding transaction: %v", err)
	}
	assert.Equal(t, txn1, txn, "encoded and decoded transaction should be identical to the original one")
}

func generateKey(t *testing.T) *ecdsa.PrivateKey {
	privateKey, err := keys.GenerateECDSA()
	if err != nil {
		t.Errorf("failed to generate a key")
	}
	return privateKey
}
