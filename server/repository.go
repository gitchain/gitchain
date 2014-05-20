package server

import (
	"encoding/hex"
	"log"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/transaction"
)

const REFUPDATE_CONFIRMATIONS_REQUIRED = 1

func RepositoryServer(srv *T) {
	ch := make(chan *block.Block)
	_, err := router.PermanentSubscribe("/block", ch)
	if err != nil {
		log.Printf("Error while subscribing to /block: %v", err)
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
					log.Printf("error during confirmation counting for %s: %v", hex.EncodeToString(tx0.Hash()), err)
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
