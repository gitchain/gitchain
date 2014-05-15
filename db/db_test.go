package db

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
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
	txn1, err := db.GetTransaction(txn.Hash())
	if err != nil {
		t.Errorf("error getting transaction: %v", err)
	}
	assert.Equal(t, txn, txn1)
}

func TestPutGetKey(t *testing.T) {
	privateKey := generateKey(t)
	key := x509.MarshalPKCS1PrivateKey(privateKey)

	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}
	err = db.PutKey("alias", key)
	if err != nil {
		t.Errorf("error putting key: %v", err)
	}
	key1 := db.GetKey("alias")
	if key1 == nil {
		t.Errorf("error getting key: %v", err)
	}
	assert.Equal(t, key, key1)
}

func generateKey(t *testing.T) *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Errorf("failed to generate a key")
	}
	return privateKey
}
