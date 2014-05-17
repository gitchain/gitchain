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
	privateKey := generateECDSAKey(t)
	txn1, rand := transaction.NewNameReservation("my-new-repository", &privateKey.PublicKey)
	txn2, _ := transaction.NewNameAllocation("my-new-repository", rand, privateKey)
	txn3, _ := transaction.NewNameDeallocation("my-new-repository", privateKey)

	transactions := []transaction.T{txn1, txn2, txn3}
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
	block1, err := db.GetBlock(block.Hash())
	if err != nil {
		t.Errorf("error getting block: %v", err)
	}
	if block1 == nil {
		t.Errorf("error getting block %v", block.Hash())
	}
	assert.Equal(t, block, block1)

	// Attempt fetching the last one
	block1, err = db.GetLastBlock()
	if err != nil {
		t.Errorf("error getting block: %v", err)
	}
	if block1 != nil {
		t.Errorf("error getting block, there should be no last block")
	}

	// Set the last one
	err = db.PutBlock(block, true)
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
	assert.Equal(t, block, block1)
}
