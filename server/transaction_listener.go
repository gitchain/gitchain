package server

import (
	"bytes"
	"time"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/types"
	"github.com/inconshreveable/log15"
)

func prepareBAT(srv *T, log log15.Logger) *transaction.Envelope {
	key, err := srv.DB.GetMainKey()
	if err != nil {
		log.Error("error while attempting to retrieve main key", "err", err)
	}
	if key != nil {
		bat, err := transaction.NewBlockAttribution()
		if err != nil {
			log.Error("error while creating a BAT", "err", err)
		} else {
			hash, err := srv.DB.GetPreviousEnvelopeHashForPublicKey(&key.PublicKey)
			if err != nil {
				log.Error("error while creating a BAT", "err", err)
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
	log := srv.Log.New("cmp", "txn")
	var msg *transaction.Envelope
	var blk *block.Block
	blockChannel := make(chan *block.Block)
	var transactionsPool []*transaction.Envelope
	var previousBlockHash types.Hash
	bch := make(chan *block.Block)
	_, err := router.PermanentSubscribe("/block", bch)
	if err != nil {
		log.Error("error while subscribing to /block", "err", err)
	}
	tch := make(chan *transaction.Envelope)
	_, err = router.PermanentSubscribe("/transaction", tch)
	if err != nil {
		log.Error("error while subscribing to /transaction", "err", err)
	}

	miningEmpty := false

initPool:
	transactionsPool = make([]*transaction.Envelope, 0)
loop:
	select {
	case msg = <-tch:
		log.Debug("received transaction", "txn", msg)
		verification, err := msg.Verify()
		if err != nil {
			log.Error("error during transaction verification", "err", err)
		}
		if !verification {
			// discard transaction
			goto loop
		}
		err = srv.DB.PutTransaction(msg)
		if err != nil {
			log.Error("error while recording transaction", "txn", msg, "err", err)
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
		if bat := prepareBAT(srv, log); bat != nil {
			blockTransactionsPool = append(transactionsPool, bat)
		}

		blk, err := block.NewBlock(previousBlockHash, targetBits(), blockTransactionsPool)
		if err != nil {
			log.Error("error while creating a new block", "err", err)
		} else {
			miningFactoryRequests <- MiningFactoryInstantiationRequest{Block: blk, ResponseChannel: blockChannel}
		}
	case blk = <-bch:
		for i := range blk.Transactions {
			err = srv.DB.DeleteTransaction(blk.Transactions[i].Hash())
			if err != nil {
				log.Error("error while deleting transaction", "txn", blk.Transactions[i], "err", err)
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
				log.Error("error while sending block", "err", err)
			}
			if err := router.Send("/block/last", make(chan *block.Block), blk); err != nil {
				log.Error("error while sending block", "err", err)
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
			if bat := prepareBAT(srv, log); bat != nil {
				transactionsPool = append(transactionsPool, bat)
			}
			blk, err := block.NewBlock(previousBlockHash, targetBits(), transactionsPool)
			if err != nil {
				log.Error("error while creating a new block", "err", err)
			} else {
				miningFactoryRequests <- MiningFactoryInstantiationRequest{Block: blk, ResponseChannel: blockChannel}
				miningEmpty = true
			}
			goto initPool

		}
	}
	goto loop
}
