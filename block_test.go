package main

import (
	"gitchain/block"
	trans "gitchain/transaction"
	"gitchain/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewBlock(t *testing.T) {
	privateKey := generateKey(t)
	txn1, rand := trans.NewNameReservation("my-new-repository", &privateKey.PublicKey)
	txn2, _ := trans.NewNameAllocation("my-new-repository", rand, privateKey)
	txn3, _ := trans.NewNameDeallocation("my-new-repository", privateKey)

	transactions := []trans.T{txn1, txn2, txn3}
	block1 := block.NewBlock(types.EmptyHash(), transactions)

	decodedTransactions := make([]trans.T, 3)
	for i := range block1.Transactions {
		decodedTransactions[i], _ = trans.Decode(block1.Transactions[i])
	}

	assert.Equal(t, transactions, decodedTransactions)
}
