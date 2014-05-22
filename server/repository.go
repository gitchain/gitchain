package server

import (
	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/transaction"
)

const REFUPDATE_CONFIRMATIONS_REQUIRED = 1

func RepositoryServer(srv *T) {
	log := srv.Log.New("cmp", "repo")
	ch := make(chan *block.Block)
	_, err := router.PermanentSubscribe("/block", ch)
	if err != nil {
		log.Error("Error while subscribing to /block", "err", err)
	}

loop:
	select {
	case blk := <-ch:
		for i := range blk.Transactions {
			tx0 := blk.Transactions[i]
			tx := tx0.Transaction
			switch tx.(type) {
			case *transaction.ReferenceUpdate:
				tx1 := tx.(*transaction.ReferenceUpdate)
				confirmations, err := srv.DB.GetTransactionConfirmations(tx0.Hash())
				if err != nil {
					log.Error("error during confirmation counting", "txn", tx0, "err", err)
					goto loop
				}
				if confirmations >= REFUPDATE_CONFIRMATIONS_REQUIRED {
					err = srv.DB.PutRef(tx1.Repository, tx1.Ref, tx1.New)
				}
			default:
				// ignore all other transactions
			}
		}
	}
	goto loop
}
