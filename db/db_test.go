package db

import (
	"bytes"
	"crypto/ecdsa"
	"errors"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/gitchain/gitchain/keys"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/types"
	"github.com/stretchr/testify/assert"
)

func TestWritableReadable(t *testing.T) {
	var e error
	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}

	writable(&e, db, func(dbtx *bolt.Tx) bool {
		bucket, _ := dbtx.CreateBucketIfNotExists([]byte("test"))
		e = bucket.Put([]byte{0}, []byte{1})
		return true
	})
	assert.Nil(t, e)

	var value []byte
	readable(&e, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("test"))
		value = bucket.Get([]byte{0})
	})
	assert.True(t, bytes.Compare(value, []byte{1}) == 0)

	writable(&e, db, func(dbtx *bolt.Tx) bool {
		bucket, _ := dbtx.CreateBucketIfNotExists([]byte("test"))
		e = bucket.Put([]byte{0}, []byte{2})
		e = errors.New("error")
		return false
	})
	assert.Equal(t, e, errors.New("error"))

	readable(&e, db, func(dbtx *bolt.Tx) {
		bucket := dbtx.Bucket([]byte("test"))
		value = bucket.Get([]byte{0})
	})
	assert.True(t, bytes.Compare(value, []byte{1}) == 0)

}

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
