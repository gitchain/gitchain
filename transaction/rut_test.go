package transaction

import (
	"testing"

	"github.com/gitchain/gitchain/util"
	"github.com/stretchr/testify/assert"
)

func TestReferenceUpdate(t *testing.T) {
	txn := NewReferenceUpdate("my-repository", "refs/heads/master", util.SHA160([]byte("random")), util.SHA160([]byte("not random")))

	assert.True(t, txn.Valid())
	txn2 := txn
	txn2.Version = 100
	assert.False(t, txn2.Valid())
	txn2 = txn
	txn2.Repository = ""
	assert.False(t, txn2.Valid())
	txn2 = txn
	txn2.Ref = ""
	assert.False(t, txn2.Valid())

}

func TestReferenceUpdateEncodingDecoding(t *testing.T) {
	txn := NewReferenceUpdate("my-repository", "refs/heads/master", util.SHA160([]byte("random")), util.SHA160([]byte("not random")))

	testTransactionEncodingDecoding(t, txn)
}
