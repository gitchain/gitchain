package server

import (
	"bytes"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/repository"
	"github.com/gitchain/gitchain/server/context"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/util"
	"github.com/inconshreveable/log15"
)

const RESERVATION_CONFIRMATIONS_REQUIRED = 3
const ALLOCATION_CONFIRMATIONS_REQUIRED = 3

func processPendingAllocations(srv *context.T, log log15.Logger) {
	pending := srv.DB.ListPendingRepositories()
	for i := range pending {
		r, err := srv.DB.GetRepository(pending[i])
		if err != nil {
			log.Error("error while processing pending repository", "repo", pending[i], "err", err)
		}
		c, err := srv.DB.GetTransactionConfirmations(r.NameAllocationTx)
		if err != nil {
			log.Error("error while calculating pending repository's allocation confirmations",
				"repo", pending[i], "txn", r.NameAllocationTx, "err", err)
		}
		if c >= ALLOCATION_CONFIRMATIONS_REQUIRED {
			r.Status = repository.ACTIVE
			srv.DB.PutRepository(r)
			log.Info("activated repository", "repo", pending[i], "alloc_txn", r.NameAllocationTx)
		}

	}
}

func isValidReservation(reservation *transaction.Envelope, alloc *transaction.Envelope) bool {
	allocT := alloc.Transaction.(*transaction.NameAllocation)
	switch reservation.Transaction.(type) {
	case *transaction.NameReservation:
		tx1 := reservation.Transaction.(*transaction.NameReservation)
		return bytes.Compare(util.SHA256(append([]byte(allocT.Name), allocT.Rand...)), tx1.Hashed) == 0
	default:
		return false
	}
}

func NameRegistrar(srv *context.T) {
	log := srv.Log.New("cmp", "name")
	ch := srv.Router.Sub("/block/last")
loop:
	select {
	case blki := <-ch:
		if blk, ok := blki.(*block.Block); ok {
			processPendingAllocations(srv, log)
			for i := range blk.Transactions {
				tx0 := blk.Transactions[i]
				tx := tx0.Transaction
				switch tx.(type) {
				case *transaction.NameAllocation:
					log.Debug("processing name allocation transaction", "txn", tx0)
					tx1 := tx.(*transaction.NameAllocation)
					// 1. find the reservation
					// 1.1. check if it was done with this server and there's a reference
					//      in scrap records
					reservation, err := srv.DB.GetScrap(util.SHA256(append(tx1.Rand, []byte(tx1.Name)...)))
					var reservationTx *transaction.Envelope
					if reservation == nil || err != nil {
						// 1.2 no scrap found, so try searching throughout database
						curBlock, err := srv.DB.GetLastBlock()
						if err != nil {
							log.Error("can't find last block during name allocation attempt")
							break
						}
						for curBlock != nil {
							for i := range curBlock.Transactions {
								if isValidReservation(curBlock.Transactions[i], tx0) {
									reservationTx = curBlock.Transactions[i]
								}
							}

							h := curBlock.PreviousBlockHash
							curBlock, err = srv.DB.GetBlock(h)
							if err != nil {
								log.Error("can't find block during name allocation attempt", "txn", h, "err", err)
								break
							}
						}

					} else {
						blk, err := srv.DB.GetTransactionBlock(reservation)
						if err != nil {
							log.Error("can't find block for name reservation", "txn", reservationTx, "err", err)
							break
						}
						for i := range blk.Transactions {
							if isValidReservation(blk.Transactions[i], tx0) {
								reservationTx = blk.Transactions[i]
								break
							}
						}

					}

					if reservationTx == nil {
						log.Error("can't find corresponding name reservation for allocation", "txn", tx0)
						break
					}

					// 2. verify its maturity
					confirmations, err := srv.DB.GetTransactionConfirmations(reservationTx.Hash())
					if err != nil {
						log.Error("can't compute number of confirmations for reservation", "txn", reservationTx.Hash(), "err", err)
						break
					}

					if confirmations >= RESERVATION_CONFIRMATIONS_REQUIRED {
						// this reservation is confirmed
						srv.DB.PutRepository(repository.NewRepository(tx1.Name, repository.PENDING, tx0.Hash()))
						log.Info("created pending repository", "repo", tx1.Name, "alloc_txn", tx0.Hash())
					} else {
						// this allocation is wasted as the distance is not long enough
					}
				default:
					// ignore all other transactions
				}
			}
		}
	}
	goto loop
}
