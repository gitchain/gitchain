package server

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/gitchain/gitchain/block"
	"github.com/gitchain/gitchain/repository"
	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/util"
)

const RESERVATION_CONFIRMATIONS_REQUIRED = 3

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

func NameRegistrar(srv *T) {
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
			case *transaction.NameAllocation:
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
						log.Printf("can't find last block during name allocation attempt")
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
							log.Printf("can't find block %s during name allocation attempt", hex.EncodeToString(h))
							break
						}
					}

				} else {
					blk, err := srv.DB.GetTransactionBlock(reservation)
					if err != nil {
						log.Printf("can't find block for name reservation %s: %v", hex.EncodeToString(reservationTx.Hash()), err)
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
					log.Printf("can't find corresponding name reservation for allocation %s", hex.EncodeToString(tx.Hash()))
					break
				}

				// 2. verify its maturity
				confirmations, err := srv.DB.GetTransactionConfirmations(reservationTx.Hash())
				if err != nil {
					log.Printf("can't compute number of confirmations for reservation %s: %v", hex.EncodeToString(reservationTx.Hash()), err)
					break
				}

				if confirmations >= RESERVATION_CONFIRMATIONS_REQUIRED {
					// this reservation is confirmed
					srv.DB.PutRepository(repository.NewRepository(tx1.Name, repository.PENDING, tx.Hash()))
				} else {
					// this allocation is wasted as the distance is not long enough
				}

			default:
				// ignore all other transactions
			}
		}
	}
	goto loop
}
