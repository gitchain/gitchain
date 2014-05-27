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

type Miner struct {
	signallingChannel chan *block.Block `json:"-"`
	responseChannel   chan *block.Block `json:"-"`
	Block             *block.Block
	StartTime         time.Time
}

type MiningStatus struct {
	Miners      []Miner
	BlocksMined int
}

func (s MiningStatus) AvailableMiners() (n int) {
	for i := range s.Miners {
		if s.Miners[i].Block == nil {
			n++
		}
	}
	return
}

type MiningFactoryStatusRequest struct {
	ResponseChannel chan MiningStatus
}

type MiningFactoryInstantiationRequest struct {
	Block           *block.Block
	ResponseChannel chan *block.Block
}

var miningFactoryRequests chan interface{} = make(chan interface{})

func GetMiningStatus() MiningStatus {
	response := make(chan MiningStatus)
	miningFactoryRequests <- &MiningFactoryStatusRequest{ResponseChannel: response}
	return <-response
}

func targetBits() uint32 {
	return 0x1f00ffff
}

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

func stopMiners(status MiningStatus) {
	for i := range status.Miners {
		status.Miners[i].signallingChannel <- nil
		status.Miners[i].Block = nil
		status.Miners[i].StartTime = time.Unix(0, 0)
	}
}

func mineBlock(status MiningStatus, srv *context.T, log log15.Logger, previousBlockHash types.Hash, transactions []*transaction.Envelope) {
	if bat := prepareBAT(srv, log); bat != nil {
		blk, err := block.NewBlock(previousBlockHash, targetBits(), append(transactions, bat))
		if err != nil {
			log.Error("error while creating a new block", "err", err)
		} else {
			// send off the new pool
			for i := range status.Miners {
				status.Miners[i].signallingChannel <- blk
				status.Miners[i].Block = blk
				status.Miners[i].StartTime = time.Now()
			}
		}
	}

}

func MiningFactory(srv *context.T) {
	log := srv.Log.New("cmp", "mining")
	var status MiningStatus
	n := 1 // TODO: in the future: srv.Config.Mining.Processes
	minedCh := make(chan *block.Block)
	for i := 0; i < n; i++ {
		ch := make(chan *block.Block)
		status.Miners = append(status.Miners, Miner{signallingChannel: ch})
		go block.Miner(ch, minedCh)
	}
	ch := srv.Router.Sub("/transaction")
	transactionsPool := make([]*transaction.Envelope, 0)
	var previousBlockHash types.Hash

	// Setup previous block hash
	if blk, _ := srv.DB.GetLastBlock(); blk == nil {
		previousBlockHash = types.EmptyHash()
	} else {
		previousBlockHash = blk.Hash()
	}

loop:
	select {
	case txni := <-ch:
		if txn, ok := txni.(*transaction.Envelope); ok {
			transactionsPool = append(transactionsPool, txn)
			stopMiners(status)
			mineBlock(status, srv, log, previousBlockHash, transactionsPool)
		}
	case blk := <-minedCh:
		if bytes.Compare(blk.PreviousBlockHash, previousBlockHash) == 0 {
			log.Debug("mined", "block", blk)
			for i := range blk.Transactions {
				for j := range transactionsPool {
					if bytes.Compare(transactionsPool[j].Hash(), blk.Transactions[i].Hash()) == 0 {
						transactionsPool = append(transactionsPool[0:j], transactionsPool[j+1:]...)
					}
				}
			}
			previousBlockHash = blk.Hash()
			stopMiners(status)
			err := srv.DB.PutBlock(blk, true)
			if err != nil {
				log.Error("error while serializing block", "block", blk.Hash(), "err", err)
			} else {
				srv.Router.Pub(blk, "/block", "/block/last")
				mineBlock(status, srv, log, previousBlockHash, transactionsPool)
			}
		}
	case <-time.After(time.Second * 1):
		if key, _ := srv.DB.GetMainKey(); key != nil && len(transactionsPool) == 0 && status.AvailableMiners() == n {
			mineBlock(status, srv, log, previousBlockHash, make([]*transaction.Envelope, 0))
		}
	}
	goto loop
}
