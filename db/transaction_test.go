package db

import (
	"os"
	"testing"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/types"
	"github.com/stretchr/testify/assert"
)

func TestGetTransactionBlock(t *testing.T) {
	transactions := fixtureSampleTransactions(t)

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

	transactions := fixtureSampleTransactions(t)
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

	anotherSampleOfTransactions := fixtureSampleTransactions(t)

	blk, err = block.NewBlock(blk.Hash(), block.HIGHEST_TARGET, anotherSampleOfTransactions)
	err = db.PutBlock(blk, true)
	if err != nil {
		t.Errorf("error putting block: %v", err)
	}

}
