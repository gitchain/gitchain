package main

import (
	"./env"
	"./transaction"
)

var ch chan transaction.T

func StartTransactionListener() {
	ch = make(chan transaction.T)
	go listener()
}

func BroadcastTransaction(tx transaction.T) {
	go func() {
		ch <- tx
	}()
}

func listener() {
	var msg transaction.T
loop:
	msg = <-ch
	env.DB.PutTransaction(msg)
	goto loop
}
