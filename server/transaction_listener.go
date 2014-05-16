package server

import (
	"bytes"
	"log"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/env"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/types"
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
	var transactionsPool []transaction.T
	var previousBlockHash types.Hash

initPool:
	transactionsPool = make([]transaction.T, 0)
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
		key, err := env.DB.GetMainKey()
		if err != nil {
			log.Printf("Error while attempting to retrieve main key: %v", err)
		}
		if key != nil {
			bat, err := transaction.NewBlockAttribution(key)
			if err != nil {
				log.Printf("Error creating a BAT")
			} else {
				transactionsPool = append(transactionsPool, bat)
			}
		} else {
			// There is no main key to create a BAT with yet
		}
		blk, err := block.NewBlock(previousBlockHash, 0x1e00ffff, transactionsPool)
		if err != nil {
			log.Printf("Error while creating a new block: %v", err)
		} else {
			miningFactoryRequests <- MiningFactoryInstantiationRequest{Block: blk, ResponseChannel: blockChannel}
		}
	case blk = <-blockChannel:
		if lastBlk, _ := env.DB.GetLastBlock(); lastBlk == nil {
			previousBlockHash = types.EmptyHash()
		} else {
			previousBlockHash = types.NewHash(blk.Hash())
		}
		isLastBlock := bytes.Compare(blk.PreviousBlockHash[:], previousBlockHash[:]) == 0
		env.DB.PutBlock(blk, isLastBlock)
		goto initPool
	}
	goto loop
}
