package db

import (
	"crypto/ecdsa"
	"testing"

	"github.com/gitchain/gitchain/keys"
	"github.com/gitchain/gitchain/transaction"
)

func fixtureSampleTransactions(t *testing.T) []transaction.T {
	privateKey := generateECDSAKey(t)
	txn1, rand := transaction.NewNameReservation("my-new-repository", &privateKey.PublicKey)
	txn2, _ := transaction.NewNameAllocation("my-new-repository", rand, privateKey)
	txn3, _ := transaction.NewNameDeallocation("my-new-repository", privateKey)

	return []transaction.T{txn1, txn2, txn3}
}

func generateECDSAKey(t *testing.T) *ecdsa.PrivateKey {
	privateKey, err := keys.GenerateECDSA()
	if err != nil {
		t.Errorf("failed to generate a key")
	}
	return privateKey
}
