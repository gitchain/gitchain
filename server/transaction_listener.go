package server

import (
	"bytes"
	"log"
	"time"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/env"
	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/types"
)

func prepareBAT() *transaction.Envelope {
	key, err := env.DB.GetMainKey()
	if err != nil {
		log.Printf("Error while attempting to retrieve main key: %v", err)
	}
	if key != nil {
		bat, err := transaction.NewBlockAttribution()
		if err != nil {
			log.Printf("Error while creating a BAT: %v", err)
		} else {
			hash, err := env.DB.GetPreviousTransactionHashForPublicKey(&key.PublicKey)
			if err != nil {
				log.Printf("Error while creating a BAT: %v", err)
			}
			bate := transaction.NewEnvelope(hash, bat)
			bate.Sign(key)
			return bate
		}
	}
	return nil

}

func targetBits() uint32 {
	return 0x1e00ffff
}

func TransactionListener() {
	var msg *transaction.Envelope
	var blk *block.Block
	blockChannel := make(chan *block.Block)
	var transactionsPool []*transaction.Envelope
	var previousBlockHash types.Hash
	ch := make(chan *transaction.Envelope)
	router.PermanentSubscribe("/transaction", ch)

	miningEmpty := false

initPool:
	transactionsPool = make([]*transaction.Envelope, 0)
loop:
	select {
	case msg = <-ch:
		verification, err := msg.Verify()
		if err != nil {
			log.Printf("error during transaction verification: %v", err)
		}
		if !verification {
			// discard transaction
			goto loop
		}
		miningEmpty = false
		transactionsPool = append(transactionsPool, msg)
		if blk, _ = env.DB.GetLastBlock(); blk == nil {
			previousBlockHash = types.EmptyHash()
		} else {
			previousBlockHash = blk.Hash()
		}
		blockTransactionsPool := make([]*transaction.Envelope, len(transactionsPool)+1)
		if bat := prepareBAT(); bat != nil {
			blockTransactionsPool = append(transactionsPool, bat)
		}
		blk, err := block.NewBlock(previousBlockHash, targetBits(), blockTransactionsPool)
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
			previousBlockHash = lastBlk.Hash()
		}
		if bytes.Compare(blk.PreviousBlockHash, previousBlockHash) == 0 {
			// this is a legitimate last block
			env.DB.PutBlock(blk, true)
			router.Send("/block", make(chan *block.Block), blk)
			router.Send("/block/last", make(chan *block.Block), blk)
		} else {
			// resubmit the block with new previous block for mining
			blk.PreviousBlockHash = previousBlockHash
			miningFactoryRequests <- MiningFactoryInstantiationRequest{Block: blk, ResponseChannel: blockChannel}
		}
		goto initPool
	case <-time.After(time.Second * 1):
		if key, _ := env.DB.GetMainKey(); len(transactionsPool) == 0 && !miningEmpty && key != nil {
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
