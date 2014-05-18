package db

import (
	"os"
	"testing"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/types"
	"github.com/stretchr/testify/assert"
)

func TestPutGetBlock(t *testing.T) {
	transactions, _ := fixtureSampleTransactions(t)

	blk, err := block.NewBlock(types.EmptyHash(), block.HIGHEST_TARGET, transactions)
	if err != nil {
		t.Errorf("can't create a block because of %v", err)
	}

	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}
	err = db.PutBlock(blk, false)
	if err != nil {
		t.Errorf("error putting block: %v", err)
	}
	block1, err := db.GetBlock(blk.Hash())
	if err != nil {
		t.Errorf("error getting block: %v", err)
	}
	if block1 == nil {
		t.Errorf("error getting block %v", blk.Hash())
	}
	assert.Equal(t, blk, block1)

	// Attempt fetching the last one
	block1, err = db.GetLastBlock()
	if err != nil {
		t.Errorf("error getting block: %v", err)
	}
	if block1 != nil {
		t.Errorf("error getting block, there should be no last block")
	}

	// Set the last one
	err = db.PutBlock(blk, true)
	if err != nil {
		t.Errorf("error putting block: %v", err)
	}
	block1, err = db.GetLastBlock()
	if err != nil {
		t.Errorf("error getting last block: %v", err)
	}
	if block1 == nil {
		t.Errorf("error getting block, there should be a last block")
	}
	assert.Equal(t, blk, block1)
}

func TestGetNextBlock(t *testing.T) {
	transactions, _ := fixtureSampleTransactions(t)

	blk, err := block.NewBlock(types.EmptyHash(), block.HIGHEST_TARGET, transactions)
	if err != nil {
		t.Errorf("can't create a block because of %v", err)
	}

	blk1, err := block.NewBlock(blk.Hash(), block.HIGHEST_TARGET, []*transaction.Envelope{})
	if err != nil {
		t.Errorf("can't create a block because of %v", err)
	}

	db, err := NewDB("test.db")
	defer os.Remove("test.db")

	if err != nil {
		t.Errorf("error opening database: %v", err)
	}

	err = db.PutBlock(blk, false)
	if err != nil {
		t.Errorf("error putting block: %v", err)
	}

	blk_, err := db.GetNextBlock(types.EmptyHash())
	if err != nil {
		t.Errorf("error getting next block: %v", err)
	}
	assert.Equal(t, blk, blk_)

	err = db.PutBlock(blk1, false)
	if err != nil {
		t.Errorf("error putting block: %v", err)
	}

	blk1_, err := db.GetNextBlock(blk.Hash())
	if err != nil {
		t.Errorf("error getting next block: %v", err)
	}
	assert.Equal(t, blk1_, blk1)

}
