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

func prepareBAT() transaction.T {
	key, err := env.DB.GetMainKey()
	if err != nil {
		log.Printf("Error while attempting to retrieve main key: %v", err)
	}
	if key != nil {
		bat, err := transaction.NewBlockAttribution(key)
		if err != nil {
			log.Printf("Error while creating a BAT: %v", err)
		} else {
			return bat
		}
	}
	return nil

}

func targetBits() uint32 {
	return 0x1e00ffff
}

func listener() {
	var msg transaction.T
	var blk *block.Block
	blockChannel := make(chan *block.Block)
	var transactionsPool []transaction.T
	var previousBlockHash types.Hash

	miningEmpty := false

initPool:
	transactionsPool = make([]transaction.T, 0)
loop:
	select {
	case msg = <-ch:
		miningEmpty = false
		env.DB.PutTransaction(msg)
		transactionsPool = append(transactionsPool, msg)
		if blk, _ = env.DB.GetLastBlock(); blk == nil {
			previousBlockHash = types.EmptyHash()
		} else {
			previousBlockHash = blk.Hash()
		}
		if bat := prepareBAT(); bat != nil {
			transactionsPool = append(transactionsPool, bat)
		}
		blk, err := block.NewBlock(previousBlockHash, targetBits(), transactionsPool)
		if err != nil {
			log.Printf("Error while creating a new block: %v", err)
		} else {
			miningFactoryRequests <- MiningFactoryInstantiationRequest{Block: blk, ResponseChannel: blockChannel}
		}
	case blk = <-blockChannel:
		miningEmpty = false
		if lastBlk, _ := env.DB.GetLastBlock(); lastBlk == nil {
			previousBlockHash = types.EmptyHash()
		} else {
			previousBlockHash = blk.Hash()
		}
		isLastBlock := bytes.Compare(blk.PreviousBlockHash, previousBlockHash) == 0
		env.DB.PutBlock(blk, isLastBlock)
		goto initPool
	default:
		if len(transactionsPool) == 0 && !miningEmpty {
			// if there are no transactions to be included into a block, try mining an empty/BAT-only block
			if blk, _ = env.DB.GetLastBlock(); blk == nil {
				previousBlockHash = types.EmptyHash()
			} else {
				previousBlockHash = blk.Hash()
			}
			if bat := prepareBAT(); bat != nil {
				transactionsPool = append(transactionsPool, bat)
			}
			blk, err := block.NewBlock(previousBlockHash, targetBits(), transactionsPool)
			if err != nil {
				log.Printf("Error while creating a new block: %v", err)
			} else {
				miningFactoryRequests <- MiningFactoryInstantiationRequest{Block: blk, ResponseChannel: blockChannel}
				miningEmpty = true
			}

		}
	}
	goto loop
}
