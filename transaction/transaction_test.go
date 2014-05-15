package transaction

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewReservation(t *testing.T) {
	privateKey := generateKey(t)
	txn, rand := NewNameReservation("my-new-repository", &privateKey.PublicKey)
	txn1, rand1 := NewNameReservation("my-new-repository", &privateKey.PublicKey)
	assert.NotEqual(t, txn.Hashed, txn1.Hashed, "hashed value should not be equal")
	assert.NotEqual(t, rand, rand1, "random numbers should not be equal")
}

func TestReservationEncodingDecoding(t *testing.T) {
	privateKey := generateKey(t)
	txn, _ := NewNameReservation("my-new-repository", &privateKey.PublicKey)

	testTransactionEncodingDecoding(t, txn)
}

func TestNewAllocation(t *testing.T) {
	privateKey := generateKey(t)
	_, rand := NewNameReservation("my-new-repository", &privateKey.PublicKey)
	txn2, err := NewNameAllocation("my-new-repository", rand, privateKey)

	if err != nil {
		t.Errorf("error while creating name allocation transaction: %v", err)
	}

	assert.True(t, txn2.Verify(&privateKey.PublicKey))

	txn2.Name = "my-old-repository"
	assert.False(t, txn2.Verify(&privateKey.PublicKey))

}

func TestAllocationEncodingDecoding(t *testing.T) {
	privateKey := generateKey(t)
	_, rand := NewNameReservation("my-new-repository", &privateKey.PublicKey)
	txn, _ := NewNameAllocation("my-new-repository", rand, privateKey)

	testTransactionEncodingDecoding(t, txn)
}

func TestNewDeallocation(t *testing.T) {
	privateKey := generateKey(t)
	txn, err := NewNameDeallocation("my-new-repository", privateKey)

	if err != nil {
		t.Errorf("error while creating name allocation transaction: %v", err)
	}

	assert.True(t, txn.Verify(&privateKey.PublicKey))

	txn.Name = "my-old-repository"
	assert.False(t, txn.Verify(&privateKey.PublicKey))

}

func TestDeallocationEncodingDecoding(t *testing.T) {
	privateKey := generateKey(t)
	txn, _ := NewNameDeallocation("my-new-repository", privateKey)
	testTransactionEncodingDecoding(t, txn)
}

func TestNewAttribution(t *testing.T) {
	privateKey := generateKey(t)
	txn, err := NewBlockAttribution(privateKey)

	if err != nil {
		t.Errorf("error while creating name allocation transaction: %v", err)
	}

	assert.True(t, txn.Verify(&privateKey.PublicKey))
	txn.Signature = []byte("boom")
	assert.False(t, txn.Verify(&privateKey.PublicKey))

}

func TestAttributionEncodingDecoding(t *testing.T) {
	privateKey := generateKey(t)
	txn, _ := NewBlockAttribution(privateKey)
	testTransactionEncodingDecoding(t, txn)
}

////

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

func generateKey(t *testing.T) *rsa.PrivateKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Errorf("failed to generate a key")
	}
	return privateKey
}
