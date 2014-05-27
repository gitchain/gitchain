package server

import (
	"github.com/gitchain/gitchain/server/context"
	"github.com/gitchain/gitchain/transaction"
)

func TransactionListener(srv *context.T) {
	log := srv.Log.New("cmp", "txn")
	tch := srv.Router.Sub("/transaction")

loop:
	select {
	case txni := <-tch:
		if txn, ok := txni.(*transaction.Envelope); ok {
			log.Debug("received transaction", "txn", txn)
			err := srv.DB.PutTransaction(txn)
			if err != nil {
				log.Error("error while recording transaction", "txn", txn, "err", err)
			}
		}
	}
	goto loop
}
