package transaction

import (
	"testing"

	"github.com/gitchain/gitchain/types"
	"github.com/stretchr/testify/assert"
)

func TestEnvelopeSignVerify(t *testing.T) {
	privateKey := generateKey(t)
	txn, _ := NewNameReservation("my-new-repository")

	e := NewEnvelope(types.EmptyHash(), txn)

	err := e.Sign(privateKey)
	if err != nil {
		t.Errorf("Can't sign the envelope: %v", err)
	}

	v, err := e.Verify()

	if err != nil {
		t.Errorf("Can't verify the envelope: %v", err)
	}

	assert.True(t, v)

}

func TestEnvelopeEncodeDecode(t *testing.T) {
	privateKey := generateKey(t)
	txn, _ := NewNameReservation("my-new-repository")

	e := NewEnvelope(types.EmptyHash(), txn)

	err := e.Sign(privateKey)
	if err != nil {
		t.Errorf("Can't sign the envelope: %v", err)
	}

	enc, err := e.Encode()
	if err != nil {
		t.Errorf("Can't encode the envelope: %v", err)
	}

	e1, err := DecodeEnvelope(enc)
	if err != nil {
		t.Errorf("Can't decode the envelope: %v", err)
	}

	assert.Equal(t, e, e1)
}
