package db

import (
	"bytes"
	"os"
	"testing"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/keys"
	"github.com/gitchain/gitchain/types"
	"github.com/stretchr/testify/assert"
)

func TestGetTransactionBlock(t *testing.T) {
	transactions, _ := fixtureSampleTransactions(t)

	block, err := block.NewBlock(types.EmptyHash(), block.HIGHEST_TARGET, transactions)
	if err != nil {
		t.Errorf("can't create a block because of %v", err)
	}

	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}
	err = db.PutBlock(block, false)
	if err != nil {
		t.Errorf("error putting block: %v", err)
	}

	// Block<->transaction indexing
	for i := range transactions {
		block1, err := db.GetTransactionBlock(transactions[i].Hash())
		if err != nil {
			t.Errorf("error getting transaction's block: %v", err)
		}
		assert.Equal(t, block, block1)
	}
}

func TestTransactionConfirmations(t *testing.T) {
	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}

	transactions, _ := fixtureSampleTransactions(t)
	confirmationsTest := func(count int, note string) {
		for i := range transactions {
			confirmations, err := db.GetTransactionConfirmations(transactions[i].Hash())
			if err != nil {
				t.Errorf("error getting transaction's confirmations: %v", err)
			}
			assert.Equal(t, confirmations, count, note)
		}
	}

	blk, err := block.NewBlock(types.EmptyHash(), block.HIGHEST_TARGET, transactions)

	if err != nil {
		t.Errorf("can't create a block because of %v", err)
	}

	confirmationsTest(0, "no transaction was confirmed yet")

	err = db.PutBlock(blk, true)
	if err != nil {
		t.Errorf("error putting block: %v", err)
	}

	confirmationsTest(1, "there should be one confirmation")

	anotherSampleOfTransactions, _ := fixtureSampleTransactions(t)

	blk, err = block.NewBlock(blk.Hash(), block.HIGHEST_TARGET, anotherSampleOfTransactions)
	err = db.PutBlock(blk, true)
	if err != nil {
		t.Errorf("error putting block: %v", err)
	}

}

func TestGetPreviousTransactionHashForPublicKey(t *testing.T) {
	transactions, _ := fixtureSampleTransactions(t)

	block, err := block.NewBlock(types.EmptyHash(), block.HIGHEST_TARGET, transactions)
	if err != nil {
		t.Errorf("can't create a block because of %v", err)
	}

	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}
	err = db.PutBlock(block, false)
	if err != nil {
		t.Errorf("error putting block: %v", err)
	}

	dec, err := keys.DecodeECDSAPublicKey(transactions[2].NextPublicKey)
	if err != nil {
		t.Errorf("error decoding ECDSA pubkey: %v", err)
	}
	tx, err := db.GetPreviousTransactionHashForPublicKey(dec)
	if err != nil {
		t.Errorf("error getting previous transaction's for a pubkey: %v", err)
	}
	assert.True(t, bytes.Compare(tx, transactions[2].Hash()) == 0)

	privateKey := generateECDSAKey(t)

	tx, err = db.GetPreviousTransactionHashForPublicKey(&privateKey.PublicKey)
	if err != nil {
		t.Errorf("error getting previous transaction's for a pubkey: %v", err)
	}
	assert.Nil(t, tx)

}

func TestGetNextTransactionHash(t *testing.T) {
	transactions, _ := fixtureSampleTransactions(t)

	block, err := block.NewBlock(types.EmptyHash(), block.HIGHEST_TARGET, transactions)
	if err != nil {
		t.Errorf("can't create a block because of %v", err)
	}

	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}
	err = db.PutBlock(block, false)
	if err != nil {
		t.Errorf("error putting block: %v", err)
	}

	tx, err := db.GetNextTransactionHash(transactions[0].Hash())
	if err != nil {
		t.Errorf("error getting next transaction: %v", err)
	}
	assert.True(t, bytes.Compare(tx, transactions[1].Hash()) == 0)

	tx, err = db.GetNextTransactionHash(transactions[1].Hash())
	if err != nil {
		t.Errorf("error getting next transaction: %v", err)
	}
	assert.True(t, bytes.Compare(tx, transactions[2].Hash()) == 0)

	tx, err = db.GetNextTransactionHash(transactions[2].Hash())
	if err != nil {
		t.Errorf("error getting next transaction: %v", err)
	}
	assert.True(t, bytes.Compare(tx, types.EmptyHash()) == 0)

}
