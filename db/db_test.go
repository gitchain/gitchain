package db

import (
	"crypto/rand"
	"crypto/rsa"
	"os"
	"testing"

	"../transaction"
	"github.com/stretchr/testify/assert"
)

func TestPutGetTransaction(t *testing.T) {
	privateKey := generateKey(t)
	txn, _ := transaction.NewNameReservation("my-new-repository", &privateKey.PublicKey)
	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}
	err = db.PutTransaction(txn)
	if err != nil {
		t.Errorf("error putting transaction: %v", err)
	}
	txn1, err := db.GetTransaction(txn.Id())
	if err != nil {
		t.Errorf("error getting transaction: %v", err)
	}
	assert.Equal(t, txn, txn1)
}

func generateKey(t *testing.T) *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Errorf("failed to generate a key")
	}
	return privateKey
}
