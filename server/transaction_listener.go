package server

import (
	"../block"
	"../env"
	"../transaction"
	"../types"
)

var ch chan transaction.T

func StartTransactionListener() {
	ch = make(chan transaction.T)
	go listener()
}

func BroadcastTransaction(tx transaction.T) {
	go func() {
		ch <- tx
	}()
}

func listener() {
	var msg transaction.T
	var blk *block.Block
	blockChannel := make(chan *block.Block)
	transactionsPool := make([]transaction.T, 0)
	var previousBlockHash types.Hash

loop:
	select {
	case msg = <-ch:
		env.DB.PutTransaction(msg)
		transactionsPool = append(transactionsPool, msg)
		if blk, _ = env.DB.GetLastBlock(); blk == nil {
			previousBlockHash = types.EmptyHash()
		} else {
			previousBlockHash = types.NewHash(blk.Hash())
		}
		blk = block.NewBlock(previousBlockHash, 0x1e00ffff, transactionsPool)
		miningFactoryRequests <- MiningFactoryInstantiationRequest{Block: blk, ResponseChannel: blockChannel}
	case newMinedBlock := <-blockChannel:
		env.DB.PutBlock(newMinedBlock, true)
		transactionsPool = make([]transaction.T, 0)
	}
	goto loop
}
