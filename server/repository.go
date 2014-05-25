package server

import (
	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/server/context"
	"github.com/gitchain/gitchain/transaction"
)

const REFUPDATE_CONFIRMATIONS_REQUIRED = 1

func RepositoryServer(srv *context.T) {
	log := srv.Log.New("cmp", "repo")
	ch := srv.Router.Sub("/block")

loop:
	select {
	case blki := <-ch:
		if blk, ok := blki.(*block.Block); ok {
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
	}
	goto loop
}
