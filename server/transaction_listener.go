package server

import (
	"bytes"
	"time"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/server/context"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/types"
	"github.com/inconshreveable/log15"
)

func prepareBAT(srv *context.T, log log15.Logger) *transaction.Envelope {
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

func TransactionListener(srv *context.T) {
	log := srv.Log.New("cmp", "txn")
	blockChannel := make(chan *block.Block)
	var transactionsPool []*transaction.Envelope
	var previousBlockHash types.Hash
	bch := srv.Router.Sub("/block")
	tch := srv.Router.Sub("/transaction")

	miningEmpty := false

initPool:
	transactionsPool = make([]*transaction.Envelope, 0)
loop:
	select {
	case txni := <-tch:
		if txn, ok := txni.(*transaction.Envelope); ok {
			log.Debug("received transaction", "txn", txn)
			verification, err := txn.Verify()
			if err != nil {
				log.Error("error during transaction verification", "err", err)
			}
			if !verification {
				// discard transaction
				goto loop
			}
			err = srv.DB.PutTransaction(txn)
			if err != nil {
				log.Error("error while recording transaction", "txn", txn, "err", err)
			}
			// notify internal components about a transaction in the working memory
			srv.Router.Pub(txn, "/transaction/mem")

			miningEmpty = false
			transactionsPool = append(transactionsPool, txn)
			if blk, _ := srv.DB.GetLastBlock(); blk == nil {
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
				miningFactoryRequests <- &MiningFactoryInstantiationRequest{Block: blk, ResponseChannel: blockChannel}
			}
		}
	case blki := <-bch:
		if blk, ok := blki.(*block.Block); ok {
			for i := range blk.Transactions {
				err := srv.DB.DeleteTransaction(blk.Transactions[i].Hash())
				if err != nil {
					log.Error("error while deleting transaction", "txn", blk.Transactions[i], "err", err)
				}
				for j := range transactionsPool {
					if bytes.Compare(transactionsPool[j].Hash(), blk.Transactions[i].Hash()) == 0 {
						transactionsPool = append(transactionsPool[0:j], transactionsPool[j+1:]...)
					}
				}
			}
		}
	case blk := <-blockChannel:
		miningEmpty = false
		if lastBlk, _ := srv.DB.GetLastBlock(); lastBlk == nil {
			previousBlockHash = types.EmptyHash()
		} else {
			previousBlockHash = lastBlk.Hash()
		}
		if bytes.Compare(blk.PreviousBlockHash, previousBlockHash) == 0 {
			// this is a legitimate last block
			for i := range blk.Transactions {
				for j := range transactionsPool {
					if bytes.Compare(transactionsPool[j].Hash(), blk.Transactions[i].Hash()) == 0 {
						transactionsPool = append(transactionsPool[0:j], transactionsPool[j+1:]...)
					}
				}
			}
			srv.DB.PutBlock(blk, true)
			srv.Router.Pub(blk, "/block", "/block/last")
			log.Debug("mined block accepted", "block", blk.Hash())
		} else {
			// resubmit the block with new previous block for mining
			log.Debug("submitting the block for re-mining", "block", blk.Hash())
			blk.PreviousBlockHash = previousBlockHash
			miningFactoryRequests <- &MiningFactoryInstantiationRequest{Block: blk, ResponseChannel: blockChannel}
		}
		goto loop
	case <-time.After(time.Second * 1):
		if key, _ := srv.DB.GetMainKey(); len(transactionsPool) == 0 && !miningEmpty && key != nil &&
			GetMiningStatus().AvailableMiners() > 0 {
			// if there are no transactions to be included into a block, try mining an empty/BAT-only block
			if blk, _ := srv.DB.GetLastBlock(); blk == nil {
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
				miningFactoryRequests <- &MiningFactoryInstantiationRequest{Block: blk, ResponseChannel: blockChannel}
				miningEmpty = true
			}
			goto initPool

		}
	}
	goto loop
}
