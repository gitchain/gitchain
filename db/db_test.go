package db

import (
	"crypto/ecdsa"
	"testing"

	"github.com/gitchain/gitchain/keys"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/types"
)

func fixtureSampleTransactions(t *testing.T) ([]*transaction.Envelope, *ecdsa.PrivateKey) {
	privateKey := generateECDSAKey(t)
	txn1, rand := transaction.NewNameReservation("my-new-repository")
	txn1e := transaction.NewEnvelope(types.EmptyHash(), txn1)
	txn1e.Sign(privateKey)
	txn2, _ := transaction.NewNameAllocation("my-new-repository", rand)
	txn2e := transaction.NewEnvelope(txn1e.Hash(), txn2)
	txn2e.Sign(privateKey)
	txn3, _ := transaction.NewNameDeallocation("my-new-repository")
	txn3e := transaction.NewEnvelope(txn2e.Hash(), txn3)
	txn3e.Sign(privateKey)

	return []*transaction.Envelope{txn1e, txn2e, txn3e}, privateKey
}

func generateECDSAKey(t *testing.T) *ecdsa.PrivateKey {
	privateKey, err := keys.GenerateECDSA()
	if err != nil {
		t.Errorf("failed to generate a key")
	}
	return privateKey
}
