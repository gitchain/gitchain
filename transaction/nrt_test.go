package transaction

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReservation(t *testing.T) {
	txn, rand := NewNameReservation("my-new-repository")
	txn1, rand1 := NewNameReservation("my-new-repository")
	assert.NotEqual(t, txn.Hashed, txn1.Hashed, "hashed value should not be equal")
	assert.NotEqual(t, rand, rand1, "random numbers should not be equal")

	assert.True(t, txn.Valid())
	txn2 := txn
	txn2.Version = 100
	assert.False(t, txn2.Valid())

}

func TestReservationEncodingDecoding(t *testing.T) {
	txn, _ := NewNameReservation("my-new-repository")

	testTransactionEncodingDecoding(t, txn)
}
