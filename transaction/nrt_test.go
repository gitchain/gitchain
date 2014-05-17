package transaction

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewReservation(t *testing.T) {
	privateKey := generateKey(t)
	txn, rand := NewNameReservation("my-new-repository", &privateKey.PublicKey)
	txn1, rand1 := NewNameReservation("my-new-repository", &privateKey.PublicKey)
	assert.NotEqual(t, txn.Hashed, txn1.Hashed, "hashed value should not be equal")
	assert.NotEqual(t, rand, rand1, "random numbers should not be equal")

	assert.True(t, txn.Valid())
	txn2 := txn
	txn2.Version = 100
	assert.False(t, txn2.Valid())

}

func TestReservationEncodingDecoding(t *testing.T) {
	privateKey := generateKey(t)
	txn, _ := NewNameReservation("my-new-repository", &privateKey.PublicKey)

	testTransactionEncodingDecoding(t, txn)
}
