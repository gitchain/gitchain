package server

import (
	"bytes"
	"log"
	"time"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/types"
)

func prepareBAT(srv *T) *transaction.Envelope {
	key, err := srv.DB.GetMainKey()
	if err != nil {
		log.Printf("Error while attempting to retrieve main key: %v", err)
	}
	if key != nil {
		bat, err := transaction.NewBlockAttribution()
		if err != nil {
			log.Printf("Error while creating a BAT: %v", err)
		} else {
			hash, err := srv.DB.GetPreviousEnvelopeHashForPublicKey(&key.PublicKey)
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
	return 0x1f00ffff
}

func TransactionListener(srv *T) {
	var msg *transaction.Envelope
	var blk *block.Block
	blockChannel := make(chan *block.Block)
	var transactionsPool []*transaction.Envelope
	var previousBlockHash types.Hash
	bch := make(chan *block.Block)
	_, err := router.PermanentSubscribe("/block", bch)
	if err != nil {
		log.Printf("Error while subscribing to /block: %v", err)
	}
	tch := make(chan *transaction.Envelope)
	_, err = router.PermanentSubscribe("/transaction", tch)
	if err != nil {
		log.Printf("Error while subscribing to /transaction: %v", err)
	}

	miningEmpty := false

initPool:
	transactionsPool = make([]*transaction.Envelope, 0)
loop:
	select {
	case msg = <-tch:
		verification, err := msg.Verify()
		if err != nil {
			log.Printf("error during transaction verification: %v", err)
		}
		if !verification {
			// discard transaction
			goto loop
		}
		err = srv.DB.PutTransaction(msg)
		if err != nil {
			log.Printf("Error while recording transaction %x: %v", msg.Hash(), err)
		}
		// notify internal components about a transaction in the working memory
		router.Send("/transaction/mem", make(chan *transaction.Envelope), msg)

		miningEmpty = false
		transactionsPool = append(transactionsPool, msg)
		if blk, _ = srv.DB.GetLastBlock(); blk == nil {
			previousBlockHash = types.EmptyHash()
		} else {
			previousBlockHash = blk.Hash()
		}
		blockTransactionsPool := make([]*transaction.Envelope, 0)
		if bat := prepareBAT(srv); bat != nil {
			blockTransactionsPool = append(transactionsPool, bat)
		}

		blk, err := block.NewBlock(previousBlockHash, targetBits(), blockTransactionsPool)
		if err != nil {
			log.Printf("Error while creating a new block: %v", err)
		} else {
			miningFactoryRequests <- MiningFactoryInstantiationRequest{Block: blk, ResponseChannel: blockChannel}
		}
	case blk = <-bch:
		for i := range blk.Transactions {
			err = srv.DB.DeleteTransaction(blk.Transactions[i].Hash())
			if err != nil {
				log.Printf("Error while deleting transaction %x: %v", blk.Transactions[i].Hash(), err)
			}
		}
	case blk = <-blockChannel:

		miningEmpty = false
		if lastBlk, _ := srv.DB.GetLastBlock(); lastBlk == nil {
			previousBlockHash = types.EmptyHash()
		} else {
			previousBlockHash = lastBlk.Hash()
		}
		if bytes.Compare(blk.PreviousBlockHash, previousBlockHash) == 0 {
			// this is a legitimate last block
			srv.DB.PutBlock(blk, true)
			if err := router.Send("/block", make(chan *block.Block), blk); err != nil {
				log.Printf("error while sending block: %v", err)
			}
			if err := router.Send("/block/last", make(chan *block.Block), blk); err != nil {
				log.Printf("error while sending block: %v", err)
			}

		} else {
			// resubmit the block with new previous block for mining
			blk.PreviousBlockHash = previousBlockHash
			miningFactoryRequests <- MiningFactoryInstantiationRequest{Block: blk, ResponseChannel: blockChannel}
		}
		goto initPool
	case <-time.After(time.Second * 1):
		if key, _ := srv.DB.GetMainKey(); len(transactionsPool) == 0 && !miningEmpty && key != nil &&
			len(GetMiningStatus().Miners) == 0 {
			// if there are no transactions to be included into a block, try mining an empty/BAT-only block
			if blk, _ = srv.DB.GetLastBlock(); blk == nil {
				previousBlockHash = types.EmptyHash()
			} else {
				previousBlockHash = blk.Hash()
			}
			if bat := prepareBAT(srv); bat != nil {
				transactionsPool = append(transactionsPool, bat)
			}
			blk, err := block.NewBlock(previousBlockHash, targetBits(), transactionsPool)
			if err != nil {
				log.Printf("Error while creating a new block: %v", err)
			} else {
				miningFactoryRequests <- MiningFactoryInstantiationRequest{Block: blk, ResponseChannel: blockChannel}
				miningEmpty = true
			}
			goto initPool

		}
	}
	goto loop
}
